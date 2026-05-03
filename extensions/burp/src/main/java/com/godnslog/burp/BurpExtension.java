package com.godnslog.burp;

import burp.api.montoya.BurpExtension;
import burp.api.montoya.MontoyaApi;
import burp.api.montoya.ui.Menu;
import burp.api.montoya.ui.editor.Editor;
import burp.api.montoya.ui.editor.extension.EditorCreationContext;
import burp.api.montoya.ui.editor.extension.EditorExtension;
import burp.api.montoya.ui.layout.ExtensionUi;
import burp.api.montoya.ui.menu.MenuItem;

/**
 * GODNSLOG Burp Suite Extension
 * 
 * Main extension entry point that registers UI components and menu items.
 */
public class BurpExtension implements BurpExtension {
    private static MontoyaApi api;
    private static String apiUrl = "http://localhost:8080/api/v2";
    private static String apiKey = "";

    @Override
    public void initialize(MontoyaApi montoyaApi) {
        BurpExtension.api = montoyaApi;
        
        // Set extension name
        montoyaApi.extension().setName("GODNSLOG OAST");
        
        // Register UI tab
        ExtensionUi ui = montoyaApi.userInterface();
        GodnslogTab tab = new GodnslogTab(montoyaApi);
        ui.registerSuiteTab("GODNSLOG", tab);
        
        // Register context menu items
        Menu menu = montoyaApi.userInterface().menu();
        menu.registerMenuItem(
            "Generate OAST Payload",
            (menuItem, event) -> {
                GodnslogPayloadGenerator generator = new GodnslogPayloadGenerator(montoyaApi);
                generator.generatePayload();
            }
        );
        
        menu.registerMenuItem(
            "Insert Payload",
            (menuItem, event) -> {
                GodnslogPayloadGenerator generator = new GodnslogPayloadGenerator(montoyaApi);
                generator.insertPayload();
            }
        );
        
        menu.registerMenuItem(
            "Monitor Interactions",
            (menuItem, event) -> {
                GodnslogTab tabInstance = tab;
                tabInstance.refreshInteractions();
            }
        );
        
        // Log initialization
        montoyaApi.logging().logToOutput("GODNSLOG OAST Extension loaded");
        montoyaApi.logging().logToOutput("API URL: " + apiUrl);
    }

    public static MontoyaApi getApi() {
        return api;
    }

    public static String getApiUrl() {
        return apiUrl;
    }

    public static void setApiUrl(String url) {
        apiUrl = url;
    }

    public static String getApiKey() {
        return apiKey;
    }

    public static void setApiKey(String key) {
        apiKey = key;
    }
}
