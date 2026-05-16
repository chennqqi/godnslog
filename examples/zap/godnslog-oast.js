// GODNSLOG OAST helper for OWASP ZAP scripting.
var apiUrl = java.lang.System.getenv("GODNSLOG_API_URL");
var apiKey = java.lang.System.getenv("GODNSLOG_API_KEY");

function authHeaders() {
  var headers = new java.util.HashMap();
  headers.put("Authorization", "Bearer " + apiKey);
  headers.put("Content-Type", "application/json");
  return headers;
}

function createPayload(caseId, template) {
  var body = JSON.stringify({ case_id: caseId, template: template, tool: "zap" });
  return org.parosproxy.paros.network.HttpSender().sendAndReceive(apiUrl + "/payloads", "POST", authHeaders(), body);
}
