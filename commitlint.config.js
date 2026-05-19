module.exports = {
  extends: ['@commitlint/config-conventional'],
  rules: {
    // Relax sentence-case to allow uppercase acronyms (HTTP, API, SQL, etc.) at start of subject.
    // All-caps (UPPER-CASE), TitleCase (start-case) and PascalCase remain forbidden.
    'subject-case': [2, 'never', ['start-case', 'pascal-case', 'upper-case']],
  },
};