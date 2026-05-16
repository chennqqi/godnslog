# xray/rad Integration

Use GODNSLOG as the private OAST evidence backend for crawler and passive scan workflows.

1. Create a Case for the scan target with `godnslog-cli case create`.
2. Generate payloads with `godnslog-cli payload create --tool xray`.
3. Inject payloads through xray/rad configuration or proxy rules.
4. Poll `godnslog-cli interaction wait --token <token>`.
5. Export evidence with `godnslog-cli report export --case-id <case> --format markdown`.
