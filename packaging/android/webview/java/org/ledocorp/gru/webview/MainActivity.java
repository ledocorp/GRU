package org.ledocorp.gru.webview;

import android.app.Activity;
import android.content.Intent;
import android.graphics.Color;
import android.net.Uri;
import android.os.Bundle;
import android.view.View;
import android.view.Window;
import android.view.WindowManager;
import android.webkit.JavascriptInterface;
import android.webkit.WebSettings;
import android.webkit.WebView;
import android.webkit.WebViewClient;

/**
 * Full-window system WebView host (Wails/Tauri-style). Go lives in libgruwebview.so
 * and talks via JNI — no raylib / NativeActivity.
 */
public class MainActivity extends Activity {
    private static final int GRU_BG = Color.rgb(228, 229, 235);

    static {
        System.loadLibrary("gruwebview");
    }

    private WebView web;

    public native void nativeStart();

    public native void nativeOnJsMessage(String msg);

    @Override
    protected void onCreate(Bundle savedInstanceState) {
        super.onCreate(savedInstanceState);
        applyLightSystemBars();

        web = new WebView(this);
        web.setBackgroundColor(GRU_BG);
        WebSettings s = web.getSettings();
        s.setJavaScriptEnabled(true);
        s.setDomStorageEnabled(true);
        s.setAllowFileAccess(true);
        web.setWebViewClient(new WebViewClient());
        web.addJavascriptInterface(new Bridge(), "gru");
        setContentView(web);

        nativeStart();
        web.loadUrl("file:///android_asset/www/index.html");
    }

    /** Called from Go to run JS on the UI thread. */
    public void evaluateJs(String script) {
        if (script == null || script.isEmpty() || web == null) {
            return;
        }
        runOnUiThread(() -> web.evaluateJavascript(script, null));
    }

    /** Optional: open external browser (same contract as NativeActivity hosts). */
    public void openGruUrl(String url) {
        if (url == null || url.isEmpty()) {
            return;
        }
        runOnUiThread(
                () -> {
                    try {
                        Intent i = new Intent(Intent.ACTION_VIEW, Uri.parse(url));
                        i.addFlags(Intent.FLAG_ACTIVITY_NEW_TASK);
                        startActivity(i);
                    } catch (Exception ignored) {
                    }
                });
    }

    @Override
    public void onBackPressed() {
        if (web != null && web.canGoBack()) {
            web.goBack();
            return;
        }
        super.onBackPressed();
    }

    private void applyLightSystemBars() {
        Window w = getWindow();
        w.addFlags(WindowManager.LayoutParams.FLAG_DRAWS_SYSTEM_BAR_BACKGROUNDS);
        w.clearFlags(WindowManager.LayoutParams.FLAG_TRANSLUCENT_STATUS);
        w.setStatusBarColor(GRU_BG);
        w.setNavigationBarColor(GRU_BG);
        View decor = w.getDecorView();
        int flags = decor.getSystemUiVisibility();
        flags |= View.SYSTEM_UI_FLAG_LIGHT_STATUS_BAR;
        decor.setSystemUiVisibility(flags);
    }

    private final class Bridge {
        @JavascriptInterface
        public void postMessage(String msg) {
            nativeOnJsMessage(msg != null ? msg : "");
        }
    }
}
