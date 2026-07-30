/**
 * gru.js — Gru WebView host bridge runtime (v1)
 *
 * Loaded by the WebView2 host at init and optionally via:
 *   <script src="https://gru.media/web/gru.js"></script>
 *
 * Contract: docs/GRU_JS.md · docs/schemas/gru-bridge-v1.json
 *
 * Design: plain static IIFE, no build step, idempotent if loaded twice.
 * Host injects this file at WebView init; modules may also script-tag it.
 */
(function (global) {
  'use strict';

  var gru = global.gru = global.gru || {};

  gru.version = 1;
  gru.capabilities = gru.capabilities || {};

  gru._handlers = gru._handlers || {};
  gru._ready = gru._ready || false;

  function postRaw(s) {
    if (global.chrome && global.chrome.webview) {
      global.chrome.webview.postMessage(s);
    }
  }

  /**
   * Post a gru.bridge envelope to Go.
   * @param {string|{type:string,v:number,name:string,payload?:object}} envelope
   */
  gru.postMessage = function (envelope) {
    if (typeof envelope === 'object') {
      try { envelope = JSON.stringify(envelope); } catch (e) { return; }
    }
    postRaw(envelope);
  };

  /**
   * @param {string} name
   * @param {function(object, object): void} fn
   */
  gru.on = function (name, fn) {
    if (!name || typeof fn !== 'function') return;
    (gru._handlers[name] = gru._handlers[name] || []).push(fn);
  };

  /**
   * @param {string} name
   * @param {function(object, object): void} [fn]
   */
  gru.off = function (name, fn) {
    var list = gru._handlers[name];
    if (!list) return;
    if (!fn) {
      delete gru._handlers[name];
      return;
    }
    gru._handlers[name] = list.filter(function (h) { return h !== fn; });
  };

  var themeTokens = {
    light: {
      '--gru-scrollbar-track': '#f1f5f9',
      '--gru-scrollbar-thumb': '#cbd5e1',
      '--gru-scrollbar-thumb-hover': '#94a3b8',
      '--gru-bg': '#ffffff',
      '--gru-fg': '#0f172a',
      '--gru-muted': '#64748b',
      '--gru-border': '#e2e8f0'
    },
    dark: {
      '--gru-scrollbar-track': '#1e293b',
      '--gru-scrollbar-thumb': '#475569',
      '--gru-scrollbar-thumb-hover': '#64748b',
      '--gru-bg': '#0f172a',
      '--gru-fg': '#e2e8f0',
      '--gru-muted': '#94a3b8',
      '--gru-border': '#334155'
    }
  };

  function injectScrollbarCSS() {
    if (!global.document) return;
    if (global.document.getElementById('gru-scrollbar-style')) return;
    var parent = global.document.head || global.document.documentElement;
    if (!parent) {
      if (!gru._scrollbarDeferred && global.document.addEventListener) {
        gru._scrollbarDeferred = true;
        global.document.addEventListener('DOMContentLoaded', function () {
          gru._scrollbarDeferred = false;
          injectScrollbarCSS();
        }, { once: true });
      }
      return;
    }
    var style = global.document.createElement('style');
    style.id = 'gru-scrollbar-style';
    style.textContent =
      '::-webkit-scrollbar { width: 10px; height: 10px; }' +
      '::-webkit-scrollbar-track { background: var(--gru-scrollbar-track, #1e293b); }' +
      '::-webkit-scrollbar-thumb { background: var(--gru-scrollbar-thumb, #475569); border-radius: 6px; }' +
      '::-webkit-scrollbar-thumb:hover { background: var(--gru-scrollbar-thumb-hover, #64748b); }';
    parent.appendChild(style);
  }

  function applyTheme(payload) {
    var doc = global.document;
    var root = doc && doc.documentElement;
    if (!root) {
      gru._pendingTheme = payload;
      if (!gru._themeDeferred && doc && doc.addEventListener) {
        gru._themeDeferred = true;
        doc.addEventListener('DOMContentLoaded', function () {
          gru._themeDeferred = false;
          if (gru._pendingTheme) {
            var pending = gru._pendingTheme;
            gru._pendingTheme = null;
            applyTheme(pending);
          }
        }, { once: true });
      }
      return;
    }
    gru._pendingTheme = null;
    var theme = (payload && payload.theme) || 'dark';
    var body = doc.body;
    root.setAttribute('data-gru-theme', theme);
    if (body) body.setAttribute('data-gru-theme', theme);
    var tokens = themeTokens[theme] || themeTokens.dark;
    Object.keys(tokens).forEach(function (key) {
      root.style.setProperty(key, tokens[key]);
    });
    if (payload && payload.tokens) {
      Object.keys(payload.tokens).forEach(function (key) {
        root.style.setProperty(key, payload.tokens[key]);
      });
    }
    injectScrollbarCSS();
  }

  gru.toast = function (text, level) {
    var doc = global.document;
    var el = doc.getElementById('gru-toast');
    if (!el) {
      el = doc.createElement('div');
      el.id = 'gru-toast';
      el.className = 'gru-toast';
      (doc.body || doc.documentElement).appendChild(el);
    }
    el.textContent = text;
    el.className = 'gru-toast show ' + (level || 'info');
    clearTimeout(gru._toastT);
    gru._toastT = setTimeout(function () { el.className = el.className.replace(/\bshow\b/g, '').trim(); }, 3200);
  };

  /**
   * FillClient only: invisible S/E/W/SE/SW grips post chrome.resize to Go.
   * Title bar (N) stays native. Enabled when capabilities.windowChromeResize.
   */
  function installWindowChromeResize() {
    var doc = global.document;
    if (!doc || gru._chromeResizeInstalled) return;
    gru._chromeResizeInstalled = true;

    var GRIP = 6;
    var CORNER = 16;
    var style = doc.createElement('style');
    style.id = 'gru-chrome-resize-style';
    style.textContent =
      '.gru-chrome-grip{position:fixed;z-index:2147483646;background:transparent;touch-action:none;}' +
      '.gru-chrome-grip[data-edge="s"]{left:0;right:0;bottom:0;height:' + GRIP + 'px;cursor:ns-resize;}' +
      '.gru-chrome-grip[data-edge="e"]{top:0;right:0;bottom:0;width:' + GRIP + 'px;cursor:ew-resize;}' +
      '.gru-chrome-grip[data-edge="w"]{top:0;left:0;bottom:0;width:' + GRIP + 'px;cursor:ew-resize;}' +
      '.gru-chrome-grip[data-edge="se"]{right:0;bottom:0;width:' + CORNER + 'px;height:' + CORNER + 'px;cursor:nwse-resize;}' +
      '.gru-chrome-grip[data-edge="sw"]{left:0;bottom:0;width:' + CORNER + 'px;height:' + CORNER + 'px;cursor:nesw-resize;}';

    function ensureDom() {
      var parent = doc.body || doc.documentElement;
      if (!parent) return false;
      if (!doc.getElementById('gru-chrome-resize-style')) {
        (doc.head || parent).appendChild(style);
      }
      if (doc.getElementById('gru-chrome-grip-se')) return true;
      ['s', 'e', 'w', 'se', 'sw'].forEach(function (edge) {
        var el = doc.createElement('div');
        el.className = 'gru-chrome-grip';
        el.id = 'gru-chrome-grip-' + edge;
        el.setAttribute('data-edge', edge);
        el.addEventListener('pointerdown', onGripDown);
        parent.appendChild(el);
      });
      return true;
    }

    var dragEdge = null;
    var moveRaf = 0;
    var lastMoveEvent = null;

    function postResize(phase, edge, e) {
      gru.postMessage({
        type: 'gru.bridge',
        v: 1,
        name: 'chrome.resize',
        payload: {
          phase: phase,
          edge: edge,
          screenX: e.screenX,
          screenY: e.screenY
        }
      });
    }

    function onGripDown(e) {
      if (e.button !== 0) return;
      var edge = e.currentTarget.getAttribute('data-edge');
      if (!edge) return;
      dragEdge = edge;
      try { e.currentTarget.setPointerCapture(e.pointerId); } catch (err) {}
      postResize('start', edge, e);
      e.preventDefault();
      e.stopPropagation();
    }

    function flushMove() {
      moveRaf = 0;
      if (!dragEdge || !lastMoveEvent) return;
      postResize('move', dragEdge, lastMoveEvent);
      lastMoveEvent = null;
    }

    function onMove(e) {
      if (!dragEdge) return;
      lastMoveEvent = e;
      if (!moveRaf) {
        moveRaf = global.requestAnimationFrame ? global.requestAnimationFrame(flushMove) : 0;
        if (!moveRaf) flushMove();
      }
    }

    function onUp(e) {
      if (!dragEdge) return;
      if (moveRaf && global.cancelAnimationFrame) {
        global.cancelAnimationFrame(moveRaf);
        moveRaf = 0;
      }
      if (lastMoveEvent) {
        postResize('move', dragEdge, lastMoveEvent);
        lastMoveEvent = null;
      }
      postResize('end', dragEdge, e);
      dragEdge = null;
    }

    if (!ensureDom()) {
      doc.addEventListener('DOMContentLoaded', function () { ensureDom(); }, { once: true });
    }
    doc.addEventListener('pointermove', onMove);
    doc.addEventListener('pointerup', onUp);
    doc.addEventListener('pointercancel', onUp);
  }

  function dispatch(raw) {
    var msg;
    try { msg = typeof raw === 'string' ? JSON.parse(raw) : raw; } catch (e) { return; }
    if (!msg || msg.type !== 'gru.bridge' || !msg.name) return;

    if (msg.name === 'theme') applyTheme(msg.payload);
    if (msg.name === 'capabilities') {
      gru.capabilities = msg.payload || {};
      if (gru.capabilities.windowChromeResize) {
        installWindowChromeResize();
      }
    }

    if (msg.name === 'pause') {
      (gru._handlers['gru.pause'] || []).forEach(function (fn) { fn(msg.payload, msg); });
    }
    if (msg.name === 'resume') {
      (gru._handlers['gru.resume'] || []).forEach(function (fn) { fn(msg.payload, msg); });
    }
    if (msg.name === 'destroy') {
      (gru._handlers['gru.destroy'] || []).forEach(function (fn) { fn(msg.payload, msg); });
    }

    (gru._handlers[msg.name] || []).forEach(function (fn) { fn(msg.payload, msg); });

    if (msg.name === 'toast' && msg.payload && msg.payload.text) {
      gru.toast(msg.payload.text, msg.payload.level || 'info');
    }
  }

  gru._dispatch = dispatch;

  /** @type {boolean} True after first init on this page. */
  gru.ready = false;

  function markReady() {
    if (gru.ready) return;
    gru.ready = true;
    (gru._handlers['gru.ready'] || []).forEach(function (fn) { fn({}, { type: 'gru.bridge', v: 1, name: 'gru.ready' }); });
  }

  if (global.chrome && global.chrome.webview) {
    global.chrome.webview.addEventListener('message', function (e) { dispatch(e.data); });
  }

  if (!gru._ready) {
    gru._ready = true;
    gru.on('theme', applyTheme);
    injectScrollbarCSS();
    markReady();
  }
})(typeof window !== 'undefined' ? window : globalThis);
