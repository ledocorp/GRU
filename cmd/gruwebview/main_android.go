//go:build android

// Android full-window WebView + Go JNI bridge (no raylib / NativeActivity).
package main

/*
#cgo LDFLAGS: -llog -landroid
#include <jni.h>
#include <stdlib.h>
#include <stdio.h>
#include <string.h>
#include <android/log.h>

#define GRU_LOG(...) __android_log_print(ANDROID_LOG_INFO, "gruwebview", __VA_ARGS__)

static JavaVM *g_vm;
static jobject g_activity; // global ref to MainActivity

static void gruEvalJS(const char *script) {
	if (g_vm == NULL || g_activity == NULL || script == NULL) {
		return;
	}
	JNIEnv *env = NULL;
	int attached = 0;
	if ((*g_vm)->GetEnv(g_vm, (void **)&env, JNI_VERSION_1_6) != JNI_OK) {
		if ((*g_vm)->AttachCurrentThread(g_vm, &env, NULL) != JNI_OK || env == NULL) {
			return;
		}
		attached = 1;
	}
	jclass cls = (*env)->GetObjectClass(env, g_activity);
	if (cls == NULL) {
		if (attached) (*g_vm)->DetachCurrentThread(g_vm);
		return;
	}
	jmethodID mid = (*env)->GetMethodID(env, cls, "evaluateJs", "(Ljava/lang/String;)V");
	if (mid != NULL) {
		jstring jscript = (*env)->NewStringUTF(env, script);
		(*env)->CallVoidMethod(env, g_activity, mid, jscript);
		if (jscript != NULL) (*env)->DeleteLocalRef(env, jscript);
	}
	(*env)->DeleteLocalRef(env, cls);
	if (attached) (*g_vm)->DetachCurrentThread(g_vm);
}

static void gruNativeStart(JNIEnv *env, jobject thiz) {
	if (g_activity != NULL) {
		(*env)->DeleteGlobalRef(env, g_activity);
		g_activity = NULL;
	}
	g_activity = (*env)->NewGlobalRef(env, thiz);
	GRU_LOG("nativeStart — Go bridge ready");
}

static void gruNativeOnJsMessage(JNIEnv *env, jobject thiz, jstring jmsg) {
	(void)thiz;
	if (jmsg == NULL) {
		return;
	}
	const char *msg = (*env)->GetStringUTFChars(env, jmsg, NULL);
	if (msg == NULL) {
		return;
	}
	GRU_LOG("js→go: %s", msg);

	size_t n = strlen(msg);
	char *buf = (char *)malloc(n * 2 + 80);
	if (buf != NULL) {
		char *p = buf;
		p += sprintf(p, "window.__gruOnGo && window.__gruOnGo(\"");
		for (size_t i = 0; i < n; i++) {
			char c = msg[i];
			if (c == '\\' || c == '"') {
				*p++ = '\\';
				*p++ = c;
			} else if (c == '\n') {
				*p++ = '\\';
				*p++ = 'n';
			} else if ((unsigned char)c < 0x20) {
				// skip controls
			} else {
				*p++ = c;
			}
		}
		strcpy(p, "\");");
		gruEvalJS(buf);
		free(buf);
	}
	(*env)->ReleaseStringUTFChars(env, jmsg, msg);
}

JNIEXPORT jint JNICALL JNI_OnLoad(JavaVM *vm, void *reserved) {
	(void)reserved;
	g_vm = vm;
	JNIEnv *env = NULL;
	if ((*vm)->GetEnv(vm, (void **)&env, JNI_VERSION_1_6) != JNI_OK || env == NULL) {
		return JNI_ERR;
	}
	jclass cls = (*env)->FindClass(env, "org/ledocorp/gru/webview/MainActivity");
	if (cls == NULL) {
		(*env)->ExceptionClear(env);
		return JNI_VERSION_1_6;
	}
	static const JNINativeMethod methods[] = {
		{"nativeStart", "()V", (void *)gruNativeStart},
		{"nativeOnJsMessage", "(Ljava/lang/String;)V", (void *)gruNativeOnJsMessage},
	};
	(*env)->RegisterNatives(env, cls, methods, (jint)(sizeof methods / sizeof methods[0]));
	(*env)->DeleteLocalRef(env, cls);
	GRU_LOG("JNI_OnLoad registered");
	return JNI_VERSION_1_6;
}
*/
import "C"

func main() {}
