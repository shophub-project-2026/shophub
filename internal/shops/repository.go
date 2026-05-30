package shops

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/shophub-project-2026/shophub/internal/metrics"
)

var ErrNotFound = errors.New("shop not found")
var ErrNameTaken = errors.New("shop name already taken")

type Repository interface {
	List(ctx context.Context, userID uuid.UUID) ([]ShopView, error)
	Create(ctx context.Context, userID uuid.UUID, in CreateInput) (*ShopView, error)
	Update(ctx context.Context, userID uuid.UUID, name string, in UpdateInput) (*ShopView, error)
	Delete(ctx context.Context, userID uuid.UUID, name string) error
}

type k8sRepository struct {
	k8s  client.Client
	pool *pgxpool.Pool
}

func NewRepository(k8s client.Client, pool *pgxpool.Pool) Repository {
	return &k8sRepository{k8s: k8s, pool: pool}
}

func (r *k8sRepository) List(ctx context.Context, userID uuid.UUID) ([]ShopView, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT shop_name, shop_namespace FROM user_shops WHERE user_id = $1`, userID)
	if err != nil {
		return nil, fmt.Errorf("query user_shops: %w", err)
	}
	defer rows.Close()

	var views []ShopView
	for rows.Next() {
		var name, ns string
		if err := rows.Scan(&name, &ns); err != nil {
			return nil, err
		}

		shop := &Shop{}
		err := r.k8s.Get(ctx, types.NamespacedName{Name: name, Namespace: ns}, shop)
		if err != nil {
			metrics.K8sOperationsTotal.WithLabelValues("get", "error").Inc()
			views = append(views, ShopView{Name: name, Namespace: ns, Phase: "Unknown"})
			continue
		}
		metrics.K8sOperationsTotal.WithLabelValues("get", "success").Inc()
		views = append(views, toView(shop))
	}
	return views, nil
}

func (r *k8sRepository) Create(ctx context.Context, userID uuid.UUID, in CreateInput) (*ShopView, error) {
	if in.Namespace == "" {
		in.Namespace = "default"
	}
	if in.Availability == "" {
		in.Availability = "standard"
	}
	if in.Database == "" {
		in.Database = "standard"
	}

	shop := &Shop{
		TypeMeta: metav1.TypeMeta{
			APIVersion: SchemeGroupVersion.String(),
			Kind:       "Shop",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      in.Name,
			Namespace: in.Namespace,
		},
		Spec: ShopSpec{
			Availability:        in.Availability,
			WalletAddress:       in.WalletAddress,
			Database:            in.Database,
			NotificationWebhook: in.NotificationWebhook,
		},
	}

	if err := r.k8s.Create(ctx, shop); err != nil {
		metrics.K8sOperationsTotal.WithLabelValues("create", "error").Inc()
		return nil, fmt.Errorf("create Shop CRD: %w", err)
	}
	metrics.K8sOperationsTotal.WithLabelValues("create", "success").Inc()

	_, err := r.pool.Exec(ctx,
		`INSERT INTO user_shops (user_id, shop_name, shop_namespace) VALUES ($1, $2, $3)`,
		userID, in.Name, in.Namespace)
	if err != nil {
		_ = r.k8s.Delete(ctx, shop)
		return nil, fmt.Errorf("register shop: %w", err)
	}

	view := toView(shop)
	return &view, nil
}

func (r *k8sRepository) Update(ctx context.Context, userID uuid.UUID, name string, in UpdateInput) (*ShopView, error) {
	ns, err := r.userNamespace(ctx, userID, name)
	if err != nil {
		return nil, err
	}

	shop := &Shop{}
	if err := r.k8s.Get(ctx, types.NamespacedName{Name: name, Namespace: ns}, shop); err != nil {
		metrics.K8sOperationsTotal.WithLabelValues("get", "error").Inc()
		return nil, ErrNotFound
	}
	metrics.K8sOperationsTotal.WithLabelValues("get", "success").Inc()

	if in.Availability != nil {
		shop.Spec.Availability = *in.Availability
	}
	if in.WalletAddress != nil {
		shop.Spec.WalletAddress = *in.WalletAddress
	}

	if err := r.k8s.Update(ctx, shop); err != nil {
		metrics.K8sOperationsTotal.WithLabelValues("update", "error").Inc()
		return nil, fmt.Errorf("update Shop CRD: %w", err)
	}
	metrics.K8sOperationsTotal.WithLabelValues("update", "success").Inc()

	view := toView(shop)
	return &view, nil
}

func (r *k8sRepository) Delete(ctx context.Context, userID uuid.UUID, name string) error {
	ns, err := r.userNamespace(ctx, userID, name)
	if err != nil {
		return err
	}

	shop := &Shop{}
	if err := r.k8s.Get(ctx, types.NamespacedName{Name: name, Namespace: ns}, shop); err != nil {
		metrics.K8sOperationsTotal.WithLabelValues("get", "error").Inc()
		return ErrNotFound
	}
	metrics.K8sOperationsTotal.WithLabelValues("get", "success").Inc()

	if err := r.k8s.Delete(ctx, shop); err != nil {
		metrics.K8sOperationsTotal.WithLabelValues("delete", "error").Inc()
		return fmt.Errorf("delete Shop CRD: %w", err)
	}
	metrics.K8sOperationsTotal.WithLabelValues("delete", "success").Inc()

	_, _ = r.pool.Exec(ctx,
		`DELETE FROM user_shops WHERE user_id = $1 AND shop_name = $2 AND shop_namespace = $3`,
		userID, name, ns)
	return nil
}

func (r *k8sRepository) userNamespace(ctx context.Context, userID uuid.UUID, name string) (string, error) {
	var ns string
	err := r.pool.QueryRow(ctx,
		`SELECT shop_namespace FROM user_shops WHERE user_id = $1 AND shop_name = $2`,
		userID, name,
	).Scan(&ns)
	if err != nil {
		return "", ErrNotFound
	}
	return ns, nil
}

func toView(s *Shop) ShopView {
	url := s.Status.ServiceURL
	if url == "" {
		// Mirror the host the shop-operator wires onto the per-shop Ingress
		// (<name>.127.0.0.1.nip.io) so the Open shop button is usable from
		// the moment the CRD lands, before the operator publishes status.
		url = "http://" + s.Name + ".127.0.0.1.nip.io"
	}
	return ShopView{
		Name:          s.Name,
		Namespace:     s.Namespace,
		Availability:  s.Spec.Availability,
		WalletAddress: s.Spec.WalletAddress,
		Database:      s.Spec.Database,
		Phase:         s.Status.Phase,
		ServiceURL:    url,
	}
}
