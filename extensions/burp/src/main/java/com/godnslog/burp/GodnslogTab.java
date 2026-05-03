package com.godnslog.burp;

import burp.api.montoya.MontoyaApi;
import burp.api.montoya.ui.UserInterface;
import burp.api.montoya.ui.component.Component;
import burp.api.montoya.ui.component.TextArea;
import burp.api.montoya.ui.layout.ExtensionUi;

import javax.swing.*;
import java.awt.*;
import java.util.Timer;
import java.util.TimerTask;

/**
 * GODNSLOG tab that displays interactions and provides controls.
 */
public class GodnslogTab implements ExtensionUi {
    private final MontoyaApi api;
    private final JPanel mainPanel;
    private final TextArea interactionList;
    private final Timer refreshTimer;

    public GodnslogTab(MontoyaApi api) {
        this.api = api;
        this.mainPanel = new JPanel(new BorderLayout());
        
        // Create interaction list
        this.interactionList = api.userInterface().createTextArea();
        
        // Create control panel
        JPanel controlPanel = new JPanel(new FlowLayout(FlowLayout.LEFT));
        
        JButton refreshButton = new JButton("Refresh");
        refreshButton.addActionListener(e -> refreshInteractions());
        controlPanel.add(refreshButton);
        
        JButton exportButton = new JButton("Export");
        exportButton.addActionListener(e -> exportInteractions());
        controlPanel.add(exportButton);
        
        JButton settingsButton = new JButton("Settings");
        settingsButton.addActionListener(e -> openSettings());
        controlPanel.add(settingsButton);
        
        // Layout
        mainPanel.add(controlPanel, BorderLayout.NORTH);
        mainPanel.add(interactionList.uiComponent(), BorderLayout.CENTER);
        
        // Auto-refresh timer
        this.refreshTimer = new Timer();
        refreshTimer.scheduleAtFixedRate(new TimerTask() {
            @Override
            public void run() {
                refreshInteractions();
            }
        }, 5000, 5000); // Refresh every 5 seconds
    }

    @Override
    public Component uiComponent() {
        return api.userInterface().createComponent(mainPanel);
    }

    public void refreshInteractions() {
        // Fetch interactions from GODNSLOG API
        GodnslogApiClient client = new GodnslogApiClient(
            BurpExtension.getApiUrl(),
            BurpExtension.getApiKey()
        );
        
        String interactions = client.listInteractions();
        
        // Update UI on EDT
        SwingUtilities.invokeLater(() -> {
            interactionList.setText(interactions);
        });
    }

    public void exportInteractions() {
        // Export interactions to file
        GodnslogApiClient client = new GodnslogApiClient(
            BurpExtension.getApiUrl(),
            BurpExtension.getApiKey()
        );
        
        String report = client.exportReport("markdown");
        
        JFileChooser fileChooser = new JFileChooser();
        fileChooser.setSelectedFile(new java.io.File("godnslog-report.md"));
        int result = fileChooser.showSaveDialog(mainPanel);
        
        if (result == JFileChooser.APPROVE_OPTION) {
            java.io.File file = fileChooser.getSelectedFile();
            try {
                java.nio.file.Files.write(file.toPath(), report.getBytes());
                api.logging().logToOutput("Report exported to: " + file.getAbsolutePath());
            } catch (Exception e) {
                api.logging().logToError("Failed to export report: " + e.getMessage());
            }
        }
    }

    public void openSettings() {
        JDialog dialog = new JDialog();
        dialog.setTitle("GODNSLOG Settings");
        dialog.setLayout(new GridLayout(3, 2, 10, 10));
        
        dialog.add(new JLabel("API URL:"));
        JTextField apiUrlField = new JTextField(BurpExtension.getApiUrl());
        dialog.add(apiUrlField);
        
        dialog.add(new JLabel("API Key:"));
        JPasswordField apiKeyField = new JPasswordField(BurpExtension.getApiKey());
        dialog.add(apiKeyField);
        
        JButton saveButton = new JButton("Save");
        saveButton.addActionListener(e -> {
            BurpExtension.setApiUrl(apiUrlField.getText());
            BurpExtension.setApiKey(new String(apiKeyField.getPassword()));
            dialog.dispose();
            api.logging().logToOutput("Settings saved");
        });
        dialog.add(saveButton);
        
        dialog.pack();
        dialog.setVisible(true);
    }
}
