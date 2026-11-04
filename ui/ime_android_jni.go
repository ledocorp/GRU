//go:build android

package ui

/*
#cgo LDFLAGS: -landroid -llog
#include <stdint.h>
#include <stdlib.h>
#include <string.h>
#include <jni.h>
#include <android/native_activity.h>
#include <android_native_app_glue.h>

extern struct android_app *GetAndroidApp(void);
extern void goImeCodepoint(int32_t cp);
extern void goImeKey(int32_t key);
extern void goImeHeight(int32_t px);
extern void goFileOpenResult(char *path, int cancelled);
extern void goFileSaveResult(char *path, int cancelled);

static void JNICALL gruNativeImeCodepoint(JNIEnv *env, jobject thiz, jint cp);
static void JNICALL gruNativeImeKey(JNIEnv *env, jobject thiz, jint key);
static void JNICALL gruNativeImeHeight(JNIEnv *env, jobject thiz, jint px);
static void JNICALL gruNativeFileOpenResult(JNIEnv *env, jobject thiz, jstring path, jint cancelled);
static void JNICALL gruNativeFileSaveResult(JNIEnv *env, jobject thiz, jstring path, jint cancelled);

// Register Java→Go natives on the live Activity class (hello / grudemo / any host).
// Deferred until android_app exists — JNI_OnLoad is too early for GetAndroidApp.
static int gru_natives_ready;

static void gruEnsureNatives(JNIEnv *env, jobject act) {
	if (gru_natives_ready || env == NULL || act == NULL) {
		return;
	}
	jclass cls = (*env)->GetObjectClass(env, act);
	if (cls == NULL) {
		return;
	}
	static const JNINativeMethod methods[] = {
		{"nativeImeCodepoint", "(I)V", (void *)gruNativeImeCodepoint},
		{"nativeImeKey", "(I)V", (void *)gruNativeImeKey},
		{"nativeImeHeight", "(I)V", (void *)gruNativeImeHeight},
		{"nativeFileOpenResult", "(Ljava/lang/String;I)V", (void *)gruNativeFileOpenResult},
		{"nativeFileSaveResult", "(Ljava/lang/String;I)V", (void *)gruNativeFileSaveResult},
	};
	if ((*env)->RegisterNatives(env, cls, methods, (jint)(sizeof methods / sizeof methods[0])) == 0) {
		gru_natives_ready = 1;
	} else {
		(*env)->ExceptionClear(env);
	}
	(*env)->DeleteLocalRef(env, cls);
}

static int gruAttachActivity(JNIEnv **envOut, jobject *actOut) {
	struct android_app *app = GetAndroidApp();
	if (app == NULL || app->activity == NULL || app->activity->vm == NULL) {
		return 0;
	}
	JNIEnv *env = NULL;
	JavaVM *vm = app->activity->vm;
	if ((*vm)->AttachCurrentThread(vm, &env, NULL) != JNI_OK || env == NULL) {
		return 0;
	}
	jobject act = app->activity->clazz;
	gruEnsureNatives(env, act);
	*envOut = env;
	*actOut = act;
	return 1;
}

static void gruCallActivityVoid(const char *name) {
	JNIEnv *env = NULL;
	jobject act = NULL;
	if (!gruAttachActivity(&env, &act)) {
		return;
	}
	jclass cls = (*env)->GetObjectClass(env, act);
	if (cls == NULL) {
		return;
	}
	jmethodID mid = (*env)->GetMethodID(env, cls, name, "()V");
	if (mid != NULL) {
		(*env)->CallVoidMethod(env, act, mid);
	}
	(*env)->DeleteLocalRef(env, cls);
}

void gruShowSoftInput(void) { gruCallActivityVoid("showGruIme"); }
void gruHideSoftInput(void) { gruCallActivityVoid("hideGruIme"); }
void gruHapticTick(void) { gruCallActivityVoid("hapticTick"); }

void gruOpenBrowser(const char *url) {
	if (url == NULL) {
		return;
	}
	JNIEnv *env = NULL;
	jobject act = NULL;
	if (!gruAttachActivity(&env, &act)) {
		return;
	}
	jclass cls = (*env)->GetObjectClass(env, act);
	if (cls == NULL) {
		return;
	}
	jmethodID mid = (*env)->GetMethodID(env, cls, "openGruUrl", "(Ljava/lang/String;)V");
	if (mid != NULL) {
		jstring jurl = (*env)->NewStringUTF(env, url);
		(*env)->CallVoidMethod(env, act, mid, jurl);
		if (jurl != NULL) {
			(*env)->DeleteLocalRef(env, jurl);
		}
	}
	(*env)->DeleteLocalRef(env, cls);
}

void gruCallPickOpen(void) {
	JNIEnv *env = NULL;
	jobject act = NULL;
	if (!gruAttachActivity(&env, &act)) {
		return;
	}
	jclass cls = (*env)->GetObjectClass(env, act);
	if (cls == NULL) {
		return;
	}
	jmethodID mid = (*env)->GetMethodID(env, cls, "gruPickOpenDocument", "()V");
	if (mid != NULL) {
		(*env)->CallVoidMethod(env, act, mid);
	}
	(*env)->DeleteLocalRef(env, cls);
}

void gruCallPickSave(const char *name, const void *data, int n) {
	JNIEnv *env = NULL;
	jobject act = NULL;
	if (!gruAttachActivity(&env, &act)) {
		return;
	}
	jclass cls = (*env)->GetObjectClass(env, act);
	if (cls == NULL) {
		return;
	}
	jmethodID mid = (*env)->GetMethodID(env, cls, "gruPickSaveDocument", "(Ljava/lang/String;[B)V");
	if (mid != NULL) {
		jstring jname = (*env)->NewStringUTF(env, name != NULL ? name : "Untitled.txt");
		jbyteArray jbytes = (*env)->NewByteArray(env, n > 0 ? n : 0);
		if (jbytes != NULL && n > 0 && data != NULL) {
			(*env)->SetByteArrayRegion(env, jbytes, 0, n, (const jbyte *)data);
		}
		(*env)->CallVoidMethod(env, act, mid, jname, jbytes);
		if (jname != NULL) {
			(*env)->DeleteLocalRef(env, jname);
		}
		if (jbytes != NULL) {
			(*env)->DeleteLocalRef(env, jbytes);
		}
	}
	(*env)->DeleteLocalRef(env, cls);
}

static void JNICALL gruNativeImeCodepoint(JNIEnv *env, jobject thiz, jint cp) {
	(void)env;
	(void)thiz;
	goImeCodepoint((int32_t)cp);
}

static void JNICALL gruNativeImeKey(JNIEnv *env, jobject thiz, jint key) {
	(void)env;
	(void)thiz;
	goImeKey((int32_t)key);
}

static void JNICALL gruNativeImeHeight(JNIEnv *env, jobject thiz, jint px) {
	(void)env;
	(void)thiz;
	goImeHeight((int32_t)px);
}

static void JNICALL gruNativeFileOpenResult(JNIEnv *env, jobject thiz, jstring path, jint cancelled) {
	(void)thiz;
	char *copy = NULL;
	if (path != NULL) {
		const char *cpath = (*env)->GetStringUTFChars(env, path, NULL);
		if (cpath != NULL) {
			copy = strdup(cpath);
			(*env)->ReleaseStringUTFChars(env, path, cpath);
		}
	}
	goFileOpenResult(copy != NULL ? copy : strdup(""), (int)cancelled);
}

static void JNICALL gruNativeFileSaveResult(JNIEnv *env, jobject thiz, jstring path, jint cancelled) {
	(void)thiz;
	char *copy = NULL;
	if (path != NULL) {
		const char *cpath = (*env)->GetStringUTFChars(env, path, NULL);
		if (cpath != NULL) {
			copy = strdup(cpath);
			(*env)->ReleaseStringUTFChars(env, path, cpath);
		}
	}
	goFileSaveResult(copy != NULL ? copy : strdup(""), (int)cancelled);
}

// Product hosts declare nativeIme* / nativeFile* on their Activity; registration
// binds to whatever class is running (see gruEnsureNatives). No hardcoded package.
JNIEXPORT jint JNICALL JNI_OnLoad(JavaVM *vm, void *reserved) {
	(void)vm;
	(void)reserved;
	return JNI_VERSION_1_6;
}
*/
import "C"

// PlatformShowSoftInput requests the Android IME via the Activity EditText proxy.
func PlatformShowSoftInput() { C.gruShowSoftInput() }

// PlatformHideSoftInput dismisses the Android IME.
func PlatformHideSoftInput() { C.gruHideSoftInput() }

// PlatformHapticTick plays a short confirmation vibration (nav change, etc.).
func PlatformHapticTick() { C.gruHapticTick() }
