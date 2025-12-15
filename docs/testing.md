# Testing Guide for Home Agent

This document provides testing procedures to ensure the Home Agent application works correctly.

## Pre-Testing Checklist

- [ ] Backend built successfully (`go build` in backend/)
- [ ] Frontend dependencies installed (`npm install` in frontend/)
- [ ] Environment variables configured (`.env` file)
- [ ] Port 8080 available (backend)
- [ ] Port 5173 available (frontend dev server)

## Unit Testing

### Frontend Type Checking

```bash
cd frontend
npm run check
```

Expected: No TypeScript errors

### Frontend Build Test

```bash
cd frontend
npm run build
```

Expected: Build succeeds, outputs to `backend/public/`

### Backend Build Test

```bash
cd backend
go build
```

Expected: Binary `home-agent` created without errors

## Integration Testing

### Test 1: Frontend Development Server

**Steps:**
1. Start backend: `cd backend && go run main.go`
2. Start frontend: `cd frontend && npm run dev`
3. Open browser to `http://localhost:5173`

**Expected Results:**
- Page loads without errors
- "Home Agent" header visible
- Connection status shows "Connected" with green indicator
- Empty state message displayed
- Input box enabled and responsive

### Test 2: WebSocket Connection

**Steps:**
1. Open browser dev tools (F12)
2. Go to Network tab, filter by WS (WebSocket)
3. Refresh page

**Expected Results:**
- WebSocket connection established to `ws://localhost:8080/ws`
- Connection status shows "Connected"
- No error messages in console

### Test 3: Send Message

**Steps:**
1. Type a message in the input box: "Hello"
2. Press Enter or click Send button

**Expected Results:**
- Message appears immediately in chat (right-aligned, blue bubble)
- Typing indicator appears (three dots animation)
- Input box is disabled during response
- Assistant response streams in (left-aligned, gray bubble)
- Connection status remains "Connected"

### Test 4: Markdown Rendering

**Steps:**
1. Send message: "Show me some code examples"
2. Wait for response with code blocks

**Expected Results:**
- Code blocks have syntax highlighting
- Markdown formatting applied (headers, lists, etc.)
- Code blocks have darker background
- Inline code styled differently from block code

### Test 5: Reconnection

**Steps:**
1. Stop the backend server (Ctrl+C)
2. Observe frontend

**Expected Results:**
- Connection status changes to "Connecting..." or "Error"
- Error banner appears: "Connection lost. Attempting to reconnect..."
- Input box disabled

**Steps continued:**
3. Restart backend server

**Expected Results:**
- Connection automatically re-establishes
- Connection status returns to "Connected"
- Error banner disappears
- Input box re-enabled

### Test 6: Long Messages

**Steps:**
1. Type a very long message (multiple paragraphs)
2. Send message

**Expected Results:**
- Textarea auto-expands during typing (up to max height)
- Message sends successfully
- Message bubbles wrap text correctly
- Scroll works properly

### Test 7: Rapid Messages

**Steps:**
1. Send multiple messages quickly
2. Observe behavior

**Expected Results:**
- All messages queued and sent
- Responses arrive in order
- No messages lost
- UI remains responsive

### Test 8: Code Block Copy

**Steps:**
1. Get a response with a code block
2. Hover over code block
3. Look for copy functionality

**Expected Results:**
- Code block clearly visible
- Able to select and copy text manually
- Syntax highlighting preserved

### Test 9: Responsive Design

**Steps:**
1. Resize browser window to mobile size (~375px width)

**Expected Results:**
- Layout adapts to narrow screen
- Header remains readable
- Messages scale appropriately (90% max-width)
- Input box remains usable
- No horizontal scrolling

### Test 10: Production Build

**Steps:**
1. Build frontend: `cd frontend && npm run build`
2. Start backend: `cd backend && ./home-agent`
3. Open browser to `http://localhost:8080`

**Expected Results:**
- Page loads from backend (not Vite)
- All functionality works same as dev mode
- WebSocket connects successfully
- Static assets load correctly

## Manual Testing Scenarios

### Scenario 1: Fresh Session

1. Clear browser storage
2. Refresh page
3. Send first message
4. Verify new session created

### Scenario 2: Long Conversation

1. Send 10+ messages
2. Verify scroll behavior
3. Check memory usage (dev tools)
4. Ensure no performance degradation

### Scenario 3: Error Handling

**Network Error:**
1. Disconnect network
2. Try to send message
3. Verify error message shown

**Invalid Message:**
1. Send empty message
2. Verify button disabled
3. Send whitespace-only message
4. Verify trimmed correctly

### Scenario 4: Keyboard Shortcuts

1. Test Enter to send
2. Test Shift+Enter for new line
3. Test Tab navigation
4. Test Escape (if applicable)

### Scenario 5: Auto-scroll

1. Send multiple messages (20+)
2. Scroll to middle of conversation
3. Send new message
4. Verify page doesn't auto-scroll (user control)
5. Scroll to bottom
6. Send new message
7. Verify page auto-scrolls

## Browser Compatibility

Test in multiple browsers:

- [ ] Chrome/Chromium (latest)
- [ ] Firefox (latest)
- [ ] Safari (latest, macOS)
- [ ] Edge (latest)

## Performance Testing

### Metrics to Check:

1. **Initial Load Time**
   - Open dev tools > Network
   - Refresh page
   - Check total load time
   - Target: < 2 seconds

2. **WebSocket Latency**
   - Send message
   - Measure time to first response chunk
   - Target: < 500ms

3. **Memory Usage**
   - Open dev tools > Memory
   - Take heap snapshot
   - Send 50 messages
   - Take another snapshot
   - Verify no major leaks

4. **Build Size**
   - Check `backend/public/assets/` after build
   - JS bundle target: < 500KB gzipped
   - CSS bundle target: < 20KB gzipped

## Debugging Tips

### Frontend Issues

```bash
# Check TypeScript errors
npm run check

# Check build process
npm run build -- --debug

# View Vite server logs
npm run dev -- --debug
```

### Backend Issues

```bash
# Run with verbose logging
go run main.go --verbose

# Check WebSocket connection
# Use wscat tool:
wscat -c ws://localhost:8080/ws
```

### Browser Console

Common things to check:
- WebSocket connection errors
- JavaScript errors
- Network requests
- Console warnings

## Known Issues

Document any known issues here:

1. **Large code blocks**: Very large code responses may cause slight lag during syntax highlighting
2. **Reconnection delay**: First reconnection attempt may take 1-2 seconds
3. **Mobile keyboard**: On mobile, keyboard may cover input box (browser behavior)

## Reporting Issues

When reporting issues, include:

1. Browser and version
2. Operating system
3. Steps to reproduce
4. Expected vs actual behavior
5. Console errors (if any)
6. Network tab screenshot (for WebSocket issues)

## Success Criteria

All tests pass when:

- [ ] Frontend builds without errors
- [ ] Backend compiles and runs
- [ ] WebSocket connects successfully
- [ ] Messages send and receive correctly
- [ ] Markdown and code highlighting work
- [ ] Reconnection works automatically
- [ ] UI is responsive on mobile
- [ ] No console errors during normal use
- [ ] Performance meets targets
