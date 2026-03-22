// Conventional Commits configuration
// https://www.conventionalcommits.org/
module.exports = {
  extends: ['@commitlint/config-conventional'],
  rules: {
    // Allowed commit types
    'type-enum': [
      2,
      'always',
      [
        'feat',     // New feature
        'fix',      // Bug fix
        'docs',     // Documentation only
        'style',    // Code style (formatting, semicolons, etc.)
        'refactor', // Code refactoring (no feature/fix change)
        'perf',     // Performance improvement
        'test',     // Adding or updating tests
        'build',    // Build system or external dependencies
        'ci',       // CI/CD configuration
        'chore',    // Maintenance tasks
        'revert',   // Revert a previous commit
        'deps',     // Dependency updates
        'release',  // Release commits
      ],
    ],
    // Subject must not be empty
    'subject-empty': [2, 'never'],
    // Type must not be empty
    'type-empty': [2, 'never'],
    // Subject max length
    'subject-max-length': [1, 'always', 100],
    // Header max length
    'header-max-length': [2, 'always', 120],
    // Body max line length (warning only)
    'body-max-line-length': [1, 'always', 200],
  },
};
