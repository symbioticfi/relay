# Contributing Guidelines

We welcome contributions from the community! Please read the following guidelines carefully to ensure a smooth development workflow and consistent project quality.

---

## Branching Strategy

We follow a simplified Git Flow model with the following primary branches:

- **`main`**: Contains stable, production-ready code. Releases are tagged from this branch.
- **`dev`**: The main development branch. All new features and fixes should be merged here first.

### Branch Naming Conventions

Please name your branches according to the purpose of the work:

- `feature/your-short-description` — for new features
- `fix/your-bug-description` — for bug fixes
- `hotfix/urgent-fix` — for critical fixes (only in special cases, coordinated with maintainers)
- `chore/your-task` — for maintenance or tooling updates
- `docs/your-doc-change` — for documentation-only changes

---

## Pull Request Process

- 🔀 **Always create pull requests targeting the `dev` branch**, never `main`.
- ✅ Make sure your branch is up to date with `dev` before opening a PR.
- ✅ Ensure your code builds and passes all tests.
- ✅ Follow Go best practices and run a linter by running `make lint`.
- 📝 Use clear and descriptive PR titles.
- 📌 Link related issues in the PR description (`Fixes #123`, `Closes #456`, etc.).

### PR Checklist

Before submitting, make sure your PR meets the following:

- [ ] Target branch is `dev`
- [ ] All tests pass
- [ ] Lint checks pass
- [ ] Code is covered with tests where applicable
- [ ] Documentation is updated if needed

---

## Releases

Releases are created by maintainers by tagging the `main` branch using semantic versioning:

- `v1.2.3` — stable releases
- `nightly-YYYYMMDD` — automated nightly builds on the `dev` branch

> 🚫 Do not tag releases manually unless you are a core maintainer.

---

## Commit Style

We encourage using [Conventional Commits](https://www.conventionalcommits.org/en/v1.0.0/) for commit messages, e.g.: