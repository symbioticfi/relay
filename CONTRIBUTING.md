# Contributing Guidelines

We welcome contributions from the community! Please read the following guidelines carefully to ensure a smooth development workflow and consistent project quality.

For detailed development setup, testing, and API change management, see the [Development Guide](DEVELOPMENT.md).

---

## Branching Strategy

We use a trunk-based development model with the following branches:

- **`main`**: The primary development branch. All active development happens here, and it contains the latest code.
- **`release-x.y`**: Release branches created for each new minor version (e.g., `release-1.0`, `release-1.1`). These branches are used to:
  - Tag stable releases (`vx.y.0`, `vx.y.1`, etc.)
  - Backport critical fixes from `main`
  - Create patch releases for older versions

### Branch Naming Conventions

Please name your branches according to the purpose of the work:

- `feature/your-short-description` — for new features
- `fix/your-bug-description` — for bug fixes
- `chore/your-task` — for maintenance or tooling updates
- `docs/your-doc-change` — for documentation-only changes

> **Note**: Release branches (`release-x.y`) are created and managed by maintainers only.

---

## Pull Request Process

- 🔀 **Always create pull requests targeting the `main` branch** (unless backporting fixes to a release branch).
- ✅ Make sure your branch is up to date with `main` (or the target `release-x.y` branch) before opening a PR.
- ✅ Ensure your code builds and passes all tests.
- ✅ Follow Go best practices and run a linter by running `make lint`.
- 📝 Use clear and descriptive PR titles.
- 📌 Link related issues in the PR description (`Fixes #123`, `Closes #456`, etc.).

> **Note**: CI workflows (tests, linting, code quality checks) run automatically on all PRs to `main` and `release-*` branches, as well as on direct commits to these branches.

### PR Checklist

Before submitting, make sure your PR meets the following:

- [ ] Target branch is `main` (or appropriate `release-x.y` branch for backports)
- [ ] All tests pass
- [ ] Lint checks pass
- [ ] Code is covered with tests where applicable
- [ ] Documentation is updated if needed

---

## Releases

Releases are created by maintainers using the following process:

1. A new `release-x.y` branch is created from `main` for each minor version (e.g., `release-1.0`, `release-1.1`)
2. Releases are tagged from the `release-x.y` branch using semantic versioning:
   - `vx.y.0` — initial release for the minor version
   - `vx.y.1`, `vx.y.2`, etc. — patch releases with backported fixes
3. Critical fixes can be cherry-picked from `main` to release branches for patch releases

This branching strategy allows us to:
- Continue active development on `main` without blocking releases
- Backport important fixes to older stable versions
- Maintain multiple supported versions simultaneously

> 🚫 Do not create release branches or tag releases manually unless you are a core maintainer.

---

## Commit Style

We encourage using [Conventional Commits](https://www.conventionalcommits.org/en/v1.0.0/) for commit messages, e.g.: