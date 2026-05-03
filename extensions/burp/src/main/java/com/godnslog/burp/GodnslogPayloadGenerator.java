package com.godnslog.burp;

import burp.api.montoya.MontoyaApi;
import burp.api.montoya.ui.contextmenu.ContextMenuEvent;

/**
 * Generates and inserts OAST payloads into Burp Suite requests.
 */
public class GodnslogPayloadGenerator {
    private final MontoyaApi api;

    public GodnslogPayloadGenerator(MontoyaApi api) {
        this.api = api;
    }

    public void generatePayload() {
        // Create payload generation dialog
        JDialog dialog = new JDialog();
        dialog.setTitle("Generate OAST Payload");
        dialog.setLayout(new GridLayout(5, 2, 10, 10));

        dialog.add(new JLabel("Payload Type:"));
        JComboBox<String> typeCombo = new JComboBox<>(new String[]{
            "SSRF", "XXE", "RFI", "RCE", "Blind SQLi", "SSTI"
        });
        dialog.add(typeCombo);

        dialog.add(new JLabel("Case ID (optional):"));
        JTextField caseIdField = new JTextField();
        dialog.add(caseIdField);

        dialog.add(new JLabel("Expiration:"));
        JComboBox<String> expireCombo = new JComboBox<>(new String[]{
            "1h", "24h", "7d", "30d"
        });
        dialog.add(expireCombo);

        dialog.add(new JLabel(""));
        JButton generateButton = new JButton("Generate");
        generateButton.addActionListener(e -> {
            String type = (String) typeCombo.getSelectedItem();
            String caseId = caseIdField.getText();
            String expiration = (String) expireCombo.getSelectedItem();

            GodnslogApiClient client = new GodnslogApiClient(
                BurpExtension.getApiUrl(),
                BurpExtension.getApiKey()
            );

            String payload = client.createPayload(type, caseId, expiration);
            
            // Show result
            JOptionPane.showMessageDialog(dialog, 
                "Payload generated:\n" + payload,
                "Success",
                JOptionPane.INFORMATION_MESSAGE);
            
            dialog.dispose();
        });
        dialog.add(generateButton);

        dialog.pack();
        dialog.setVisible(true);
    }

    public void insertPayload() {
        // Get selected text from current editor
        // This is a simplified version - actual implementation would need
        // to get the current editor context
        String selectedText = "{{.Token}}.yourdomain.com";
        
        // Generate payload
        GodnslogApiClient client = new GodnslogApiClient(
            BurpExtension.getApiUrl(),
            BurpExtension.getApiKey()
        );
        
        String payload = client.createPayload("generic", "", "24h");
        
        // Replace selected text with payload
        // This would need to interact with the current editor
        api.logging().logToOutput("Payload generated: " + payload);
        api.logging().logToOutput("Insert payload into selected text (not implemented in MVP)");
    }
}
