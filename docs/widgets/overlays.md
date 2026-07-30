# Overlays (modals, toasts, tabs, tooltips)

**Purpose:** Overlay widgets float above the main document tree — blocking dialogs, transient toasts, tabbed sections, and hover tooltips. They use package-level managers or wrapper nodes; integrate them in the main loop explicitly.

**Related:** [../COMPOSITION.md](../COMPOSITION.md) · [buttons-signals.md](buttons-signals.md)

---

## Key types

### Modal

```go
body := ui.NewLabel("confirm-msg", "Delete this item?", 0, 0, 0, 0)
ui.ShowModal("Confirm", body, []ui.ModalButton{
    {Label: "Delete", Action: func() { delete(); ui.CloseModal() }},
    {Label: "Cancel", Action: ui.CloseModal},
})

// Sized variant:
ui.ShowModalSized("Edit", formNode, buttons, 520, 400)

// Main loop (after doc.Root.Update):
ui.ModalMgr.Update(dt)
// Inside BeginDrawing, after widget draws:
ui.ModalMgr.Draw()
```

| API | Role |
|-----|------|
| `ShowModal(title, content Node, buttons []ModalButton)` | Standard centered dialog |
| `ShowModalSized(..., w, h float32)` | Explicit box size |
| `CloseModal()` | Dismiss with animation |
| `ModalMgr` | Singleton update/draw host |

**Rule:** modal content is a **flat flex body** — do **not** use a `Viewport` as the modal root.

### Toast

```go
ui.ShowToast("Saved", ui.ToastSuccess, 2*time.Second)
ui.ShowToastWithAction("Updated", ui.ToastInfo, 3*time.Second, "Undo", undoFn)

// Main loop:
ui.Toasts.Update(dt)
ui.Toasts.Draw()
```

| API | Role |
|-----|------|
| `ShowToast(message, level ToastLevel, duration)` | Auto-dismiss banner |
| `ShowToastWithAction(..., actionLabel, onAction)` | Trailing action chip |
| `ToastInfo`, `ToastSuccess`, `ToastWarning`, `ToastError` | Level constants |

Batch 7 also demonstrates notification history — see scene for center panel patterns.

### TabView

```go
tabs := ui.NewTabView("settings-tabs", 0, 0, 0, 0)
tabs.AddTab("General", generalPanel)
tabs.AddTab("Advanced", advancedPanel)
```

| Constructor | `NewTabView(id string, x, y, w, h float32) *TabView` |

Each tab hosts a content node; tab bar handles selection and keyboard focus within the widget.

### Tooltip

```go
tip := ui.NewTooltip("help-save", saveButton, "Save changes")
row.AddChild(tip) // add Tooltip, not the target alone
```

| Constructor | `NewTooltip(id string, target Node, text string) *Tooltip` |

The tooltip wraps the target: pass the **`Tooltip`** to the parent container. `Delay` defaults to 0.4s hover.

---

## Demo links

| Goal | Scene title (exact) |
|------|---------------------|
| Tooltip, TabView, Modal together | **Batch 1 · Tooltip / TabView / Modal** |
| Toasts + notification center | **Batch 7 · Toast / Notification** |
| Tabbed sections in a panel | **Batch 1 · Tooltip / TabView / Modal** |

---

## Pitfalls

| Mistake | Fix |
|---------|-----|
| `Viewport` as modal root | Flat flex column of labels/inputs only |
| Forgetting `ModalMgr.Update` / `Draw` | Wire into main loop |
| Forgetting `Toasts.Update` / `Draw` | Same as modals |
| Adding target widget directly after `NewTooltip` | Add the `Tooltip` wrapper node |
| Stacking multiple blocking modals without closing | Close or size explicitly with `ShowModalSized` |

---

## Composition notes

1. **Modals are app-global** — content node can be any `Node` tree, but keep it shallow (labels, buttons, small forms).
2. **Toasts wake overlay FPS** — short-lived; do not rely on them for critical confirmation (use modal for destructive actions).
3. **Tabs inside panels:** Batch 1 mounts `TabView` inside a `Panel` on a scroll page — good default for settings-style overlays.
4. **Tooltips on icon buttons:** wrap `IconButton` targets for compact chrome.
5. **Theme keys:** `modal`, `modal-title`, `tooltip`; toast styles are internal to the toast manager.
