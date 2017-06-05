package edit

import (
	//"log"
	"reflect"
)

// NOTE: idea for custom go motion: like j or k but combine them with an action, 3Md deletes 3 lines above and 3 lines below

type Mode int
type ParseActionState int
type ParseFunc func(*Action) ParseActionState
type MotionFunc func(*Vim, *Action, Buffer) Range
type VerbFunc func(*Vim, Buffer, Range) error

const (
	MODE_NORMAL Mode = iota
	MODE_INSERT
	MODE_VISUAL_RANGE
	MODE_VISUAL_LINE
	MODE_VISUAL_BLOCK
	MODE_REPLACE
)

const (
	PARSE_ACTION_STATE_INVALID ParseActionState = iota
	PARSE_ACTION_STATE_KEY_NOT_HANDLED
	PARSE_ACTION_STATE_CONSUME_ADDITIONAL_KEY
	PARSE_ACTION_STATE_IN_PROGRESS
	PARSE_ACTION_STATE_COMPLETE
)

type KeyBind struct {
	function ParseFunc
	key      rune
}

type Verb struct {
	function VerbFunc
	param    string
}

type Motion struct {
	function   MotionFunc
	multiplier int
	param      string
}

type Action struct {
	multiplier int
	motion     Motion
	verb       Verb
	final_mode Mode
	yank       bool
}

type Vim struct {
	mode    Mode
	command []rune
	binds   []KeyBind
}

type Range struct {
	start Point
	end   Point
}

type Span struct {
	start int
	end   int
}

func (vim *Vim) Init() {
	vim.binds = append(vim.binds, KeyBind{key: 'h', function: parseMotionLeft})
	vim.binds = append(vim.binds, KeyBind{key: 'l', function: parseMotionRight})
	vim.binds = append(vim.binds, KeyBind{key: 'j', function: parseMotionDown})
	vim.binds = append(vim.binds, KeyBind{key: 'k', function: parseMotionUp})
	vim.binds = append(vim.binds, KeyBind{key: 'd', function: parseVerbDelete})
}

func (vim *Vim) ParseAction(key rune) (state ParseActionState, action Action) {
	action.multiplier = 1
	action.motion.multiplier = 1
	vim.command = append(vim.command, key)

	// parse the commands
	for _, command_key := range vim.command {
		state = PARSE_ACTION_STATE_INVALID

		for _, bind := range vim.binds {
			if bind.key == command_key {
				state = bind.function(&action)

				switch state {
				default:
				case PARSE_ACTION_STATE_INVALID:
				case PARSE_ACTION_STATE_COMPLETE:
					vim.command = []rune{}
					return state, action
				}

				break
			}
		}

		if state == PARSE_ACTION_STATE_INVALID {
			vim.command = []rune{}
			return state, action
		}
	}

	return state, action
}

func (vim *Vim) Perform(action *Action, buffer Buffer) (err error) {
	r := action.motion.function(vim, action, buffer)
	return action.verb.function(vim, buffer, r)
}

func (r *Range) Sort() {
	if r.start.IsAfter(r.end) {
		tmp := r.start
		r.start = r.end
		r.end = tmp
	}
}

// parse functions
func parseMotionLeft(action *Action) ParseActionState {
	action.motion.function = motionLeft
	if action.verb.function == nil {
		action.verb.function = verbMotion
	}
	return PARSE_ACTION_STATE_COMPLETE
}

func parseMotionRight(action *Action) ParseActionState {
	action.motion.function = motionRight
	if action.verb.function == nil {
		action.verb.function = verbMotion
	}
	return PARSE_ACTION_STATE_COMPLETE
}

func parseMotionUp(action *Action) ParseActionState {
	action.motion.function = motionUp
	if action.verb.function == nil {
		action.verb.function = verbMotion
	}
	return PARSE_ACTION_STATE_COMPLETE
}

func parseMotionDown(action *Action) ParseActionState {
	action.motion.function = motionDown
	if action.verb.function == nil {
		action.verb.function = verbMotion
	}
	return PARSE_ACTION_STATE_COMPLETE
}

func parseVerbDelete(action *Action) ParseActionState {
	if action.verb.function == nil {
		action.verb.function = verbDelete
		return PARSE_ACTION_STATE_IN_PROGRESS
	} else {
		action.motion.function = motionCurrentLine
	}

	return PARSE_ACTION_STATE_COMPLETE
}

// motion functions
func motionLeft(vim *Vim, action *Action, buffer Buffer) (r Range) {
	r.start = buffer.Cursor()
	r.end = MoveCursor(buffer, r.start, Point{-1, 0})
	return r
}

func motionRight(vim *Vim, action *Action, buffer Buffer) (r Range) {
	r.start = buffer.Cursor()
	r.end = MoveCursor(buffer, r.start, Point{1, 0})
	return r
}

func motionUp(vim *Vim, action *Action, buffer Buffer) (r Range) {
	r.start = buffer.Cursor()
	aF := reflect.ValueOf(verbMotion)
	bF := reflect.ValueOf(action.verb.function)
	if aF.Pointer() == bF.Pointer() {
		r.end = MoveCursor(buffer, r.start, Point{0, -1})
	} else {
		r.start.X = stringLastIndex(buffer.Lines()[r.start.Y])
		r.end.Y = r.start.Y - 1
		r.end.X = 0
	}
	return r
}

func motionDown(vim *Vim, action *Action, buffer Buffer) (r Range) {
	r.start = buffer.Cursor()
	aF := reflect.ValueOf(verbMotion)
	bF := reflect.ValueOf(action.verb.function)
	if aF.Pointer() == bF.Pointer() {
		r.end = MoveCursor(buffer, r.start, Point{0, 1})
	} else {
		r.start.X = 0
		r.end.Y = r.start.Y + 1
		r.end.X = stringLastIndex(buffer.Lines()[r.end.Y])
	}
	return r
}

func motionCurrentLine(vim *Vim, action *Action, buffer Buffer) (r Range) {
	line := buffer.Cursor().Y
	r.start = Point{0, line}
	r.end = Point{stringLastIndex(buffer.Lines()[line]), line}
	return r
}

// verb functions
func verbMotion(vim *Vim, buffer Buffer, r Range) (err error) {
	if vim.mode != MODE_INSERT {
		r.end = ClampIn(buffer, r.end)
	}

	err = buffer.SetCursor(r.end)
	return
}

func verbDelete(vim *Vim, buffer Buffer, r Range) (err error) {
	var spans []Span

	// calculate where the cursor will end, don't move it unless we are deleting up
	end_cursor := buffer.Cursor()
	if end_cursor.IsAfter(r.end) {
		end_cursor.Y = r.end.Y
		if r.start.Y == r.end.Y {
			end_cursor.X = r.end.X
		}
	}

	r.Sort()

	// find line spans
	for l := r.start.Y; l <= r.end.Y; l++ {
		var span Span

		// if we are on the starting line, use start.X, otherwise use the beginning of the line
		if l == r.start.Y {
			span.start = r.start.X
		} else {
			span.start = 0
		}

		// if we are on the ending line, use end.X, otherwise use the end of the line
		if l == r.end.Y {
			span.end = r.end.X
		} else {
			span.end = stringLastIndex(buffer.Lines()[l])
		}

		spans = append(spans, span)
	}

	// set line or delete line for each span
	deleted_lines := 0
	for i, span := range spans {
		line_index := r.start.Y + i - deleted_lines

		if span.start == 0 && span.end == stringLastIndex(buffer.Lines()[line_index]) {
			// the range included the entire line, so just remove it
			DeleteLine(buffer, line_index)
			deleted_lines += 1
		} else {
			// just delete a range of characters within the line
			line := buffer.Lines()[line_index]
			new_line := line[0:span.start] + line[span.end:len(line)]
			SetLine(buffer, line_index, new_line)
		}
	}

	// update the cursor
	buffer.SetCursor(ClampIn(buffer, end_cursor))
	return
}

// helpers
func stringLastIndex(str string) (index int) {
	result := len(str)
	if result > 0 {
		result -= 1
	}
	return result
}
