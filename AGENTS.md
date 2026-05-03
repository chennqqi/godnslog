# Repository Guidelines

## Project Structure & Module Organization

GODNSLOG is a Go application with a Vue 2 frontend. Root Go entry points live in `main.go`, `servecmd.go`, and `resetpass.go`. Backend server code is in `server/`, with generated Swagger output under `server/docs/`. Database models are in `models/`, cache logic is in `cache/`, and SDK/sample clients are in `client/` and `examples/`. The Vue CLI frontend lives in `frontend/`; source files are under `frontend/src/`, static files under `frontend/public/`, and UI assets under `frontend/src/assets/`. Documentation and screenshots are kept in `doc/` and `res/`.

## Build, Test, and Development Commands

- `go build`: builds the backend binary from the repository root.
- `go test ./...`: runs all Go tests.
- `go run . serve -domain example.com -4 127.0.0.1`: starts the service locally.
- `cd frontend && yarn install`: installs frontend dependencies.
- `cd frontend && yarn serve`: runs the Vue development server.
- `cd frontend && yarn build`: creates the production frontend bundle.
- `cd frontend && yarn lint`: runs the Vue/JavaScript linter.
- `docker build -t user/godnslog .`: builds the container image; use `-f DockerfileCN` for the China-oriented Dockerfile.

## Coding Style & Naming Conventions

Format Go code with `gofmt`; keep package names short, lowercase, and aligned with directory names. Use exported Go identifiers only for API surface consumed across packages. Frontend code follows Vue CLI Standard/ESLint conventions: two-space indentation, PascalCase single-file components, and route/view files grouped by feature under `frontend/src/views/`.

## Testing Guidelines

Add Go tests next to the code they cover using `*_test.go` files and `TestXxx` function names. Run `go test ./...` before submitting backend changes. For frontend behavior, use the configured Jest runner with `cd frontend && yarn test:unit`; add tests near the component or module being changed when practical.

## Commit & Pull Request Guidelines

Recent history uses short subjects such as `update doc` and `fix logout confirm cover bug`, plus standard merge commits. Keep commit titles concise and focused on one change. Pull requests should include a summary, tests performed, linked issues when applicable, and screenshots or GIFs for visible frontend changes.

## Security & Configuration Tips

Do not commit real domains, API tokens, production database files, or generated credentials. The default admin password is printed on first run and can be changed with `resetpw`. Treat DNS, callback, and rebinding configuration as sensitive.
