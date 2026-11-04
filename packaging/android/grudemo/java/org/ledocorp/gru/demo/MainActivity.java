package org.ledocorp.gru.demo;

import android.app.NativeActivity;
import android.content.Intent;
import android.graphics.Color;
import android.net.Uri;
import android.os.Bundle;
import android.view.View;
import android.view.Window;
import android.view.WindowManager;

/**
 * Colors the system status/nav bars to match Gru's clear color so the OS
 * chrome band is not a stark black slab. No LAYOUT_FULLSCREEN / immersive —
 * those either blanked GLES or doubled the top inset against mAppBounds.
 */
public class MainActivity extends NativeActivity {
    private static final int GRU_BG = Color.rgb(228, 229, 235);

    @Override
    protected void onCreate(Bundle savedInstanceState) {
        super.onCreate(savedInstanceState);
        applyLightSystemBars();
    }

    @Override
    public void onWindowFocusChanged(boolean hasFocus) {
        super.onWindowFocusChanged(hasFocus);
        if (hasFocus) {
            applyLightSystemBars();
        }
    }

    /** Called from Go OpenBrowser — launch system browser / Custom Tabs handler. */
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
}
