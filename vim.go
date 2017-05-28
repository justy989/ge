package main

// NOTE: idea for custom go motion: like j or k but combine them with an action, 3Md deletes 3 lines above and 3 lines below

type Mode int
type ParseActionState int
type ParseFunc func(*Action) ParseActionState
type MotionFunc func(*Vim, *Action, Buffer) (Point, Point)
type VerbFunc func(*Vim, Buffer, Point, Point) error

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

func (vim *Vim) init() {
	vim.binds = append(vim.binds, KeyBind{key: 'h', function: parseMotionLeft})
	vim.binds = append(vim.binds, KeyBind{key: 'l', function: parseMotionRight})
	vim.binds = append(vim.binds, KeyBind{key: 'j', function: parseMotionDown})
	vim.binds = append(vim.binds, KeyBind{key: 'k', function: parseMotionUp})
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
			}
		}
	}

	return state, action
}

func (vim *Vim) Perform(action *Action, buffer Buffer) (err error) {
	start, end := action.motion.function(vim, action, buffer)
	return action.verb.function(vim, buffer, start, end)
}

// parse functions
func parseMotionLeft(action *Action) ParseActionState {
	action.motion.function = motionLeft
	action.verb.function = verbMotion
	return PARSE_ACTION_STATE_COMPLETE
}

func parseMotionRight(action *Action) ParseActionState {
	action.motion.function = motionRight
	action.verb.function = verbMotion
	return PARSE_ACTION_STATE_COMPLETE
}

func parseMotionUp(action *Action) ParseActionState {
	action.motion.function = motionUp
	action.verb.function = verbMotion
	return PARSE_ACTION_STATE_COMPLETE
}

func parseMotionDown(action *Action) ParseActionState {
	action.motion.function = motionDown
	action.verb.function = verbMotion
	return PARSE_ACTION_STATE_COMPLETE
}

// motion functions
func motionLeft(vim *Vim, action *Action, buffer Buffer) (start Point, end Point) {
	start = buffer.Cursor()
	end = MoveCursor(buffer, start, Point{-1, 0})
	return start, end
}

func motionRight(vim *Vim, action *Action, buffer Buffer) (start Point, end Point) {
	start = buffer.Cursor()
	end = MoveCursor(buffer, start, Point{1, 0})
	return start, end
}

func motionUp(vim *Vim, action *Action, buffer Buffer) (start Point, end Point) {
	start = buffer.Cursor()
	end = MoveCursor(buffer, start, Point{0, -1})
	return start, end
}

func motionDown(vim *Vim, action *Action, buffer Buffer) (start Point, end Point) {
	start = buffer.Cursor()
	end = MoveCursor(buffer, start, Point{0, 1})
	return start, end
}

// verb functions
func verbMotion(vim *Vim, buffer Buffer, start Point, end Point) error {
	buffer.SetCursor(end)
	return nil
}
