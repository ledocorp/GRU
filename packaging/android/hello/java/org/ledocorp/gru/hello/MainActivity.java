package org.ledocorp.gru.hello;

import android.app.NativeActivity;
import android.content.Intent;
import android.net.Uri;

/** Thin NativeActivity subclass — see grudemo MainActivity notes. */
public class MainActivity extends NativeActivity {

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
}
