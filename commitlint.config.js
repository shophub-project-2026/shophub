module.exports = {
  extends: ['@commitlint/config-conventional'],
  ignores: [
    (commit) => commit.startsWith('Merge branch'),
    (commit) => commit.startsWith('Merge pull request'),
    (commit) => /\(#\d+\)/.test(commit.split('\n')[0]),
  ],
  rules: {
    'subject-case': [2, 'never', ['start-case', 'pascal-case', 'upper-case']],
  },
};