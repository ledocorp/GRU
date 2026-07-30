package ui

// MinClientWidth is the smallest logical client width the demo launcher and
// borderless chrome enforce. OS resize below this is clamped back up with
// SetWindowSize so layouts never run in an unusably narrow strip (see
// ARCHITECTURE.md — minimum width & expansion contract).
// 480px matches BreakpointXS; mobile chrome still stacks below that tier.
const MinClientWidth int32 = 480
