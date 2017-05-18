package main

type Mode int
type VerbType int
type MotionType int
type ParseActionState int
type VimKeyFunc func(*Action) ParseActionState

const (
	MODE_NORMAL Mode = iota
	MODE_INSERT
	MODE_VISUAL_RANGE
	MODE_VISUAL_LINE
	MODE_VISUAL_BLOCK
	MODE_REPLACE
)

const (
	VERB_TYPE_NONE VerbType = iota
	VERB_TYPE_MOTION
	VERB_TYPE_DELETE
	VERB_TYPE_CHANGE_CHAR
	VERB_TYPE_PASTE_BEFORE
	VERB_TYPE_PASTE_AFTER
	VERB_TYPE_YANK
	VERB_TYPE_INDENT
	VERB_TYPE_UNINDENT
	VERB_TYPE_COMMENT
	VERB_TYPE_UNCOMMENT
	VERB_TYPE_FLIP_CASE
	VERB_TYPE_JOIN_LINE
	VERB_TYPE_OPEN_ABOVE
	VERB_TYPE_OPEN_BELOW
	VERB_TYPE_SET_MARK
	VERB_TYPE_RECORD_MACRO
	VERB_TYPE_PLAY_MACRO
	VERB_TYPE_SUBSTITUTE
)

const (
	MOTION_TYPE_NONE MotionType = iota
	MOTION_TYPE_LEFT
	MOTION_TYPE_RIGHT
	MOTION_TYPE_UP
	MOTION_TYPE_DOWN
	MOTION_TYPE_WORD_LITTLE
	MOTION_TYPE_WORD_BIG
	MOTION_TYPE_WORD_BEGINNING_LITTLE
	MOTION_TYPE_WORD_BEGINNING_BIG
	MOTION_TYPE_WORD_END_LITTLE
	MOTION_TYPE_WORD_END_BIG
	MOTION_TYPE_LINE
	MOTION_TYPE_LINE_UP
	MOTION_TYPE_LINE_DOWN
	MOTION_TYPE_FIND_NEXT_MATCHING_CHAR
	MOTION_TYPE_FIND_PREV_MATCHING_CHAR
	MOTION_TYPE_LINE_SOFT
	MOTION_TYPE_TO_NEXT_MATCHING_CHAR
	MOTION_TYPE_TO_PREV_MATCHING_CHAR
	MOTION_TYPE_BEGINNING_OF_FILE
	MOTION_TYPE_BEGINNING_OF_LINE_HARD
	MOTION_TYPE_BEGINNING_OF_LINE_SOFT
	MOTION_TYPE_END_OF_LINE_PASSED
	MOTION_TYPE_END_OF_LINE_HARD
	MOTION_TYPE_END_OF_LINE_SOFT
	MOTION_TYPE_END_OF_FILE
	MOTION_TYPE_INSIDE_PAIR
	MOTION_TYPE_INSIDE_WORD_LITTLE
	MOTION_TYPE_INSIDE_WORD_BIG
	MOTION_TYPE_AROUND_PAIR
	MOTION_TYPE_AROUND_WORD_LITTLE
	MOTION_TYPE_AROUND_WORD_BIG
	MOTION_TYPE_VISUAL_RANGE
	MOTION_TYPE_VISUAL_LINE
	MOTION_TYPE_VISUAL_SWAP_WITH_CURSOR
	MOTION_TYPE_SEARCH_WORD_UNDER_CURSOR
	MOTION_TYPE_SEARCH
	MOTION_TYPE_MATCHING_PAIR
	MOTION_TYPE_NEXT_BLANK_LINE
	MOTION_TYPE_PREV_BLANK_LINE
	MOTION_TYPE_GOTO_MARK
)

const (
	PARSE_ACTION_STATE_INVALID ParseActionState = iota
	PARSE_ACTION_STATE_IN_PROGRESS
	PARSE_ACTION_STATE_COMPLETE
)

type VimKey struct {
	function VimKeyFunc
	key      rune
}

type Verb struct {
	verb_type   VerbType
	string_data string
	int_data    int
}

type Motion struct {
	motion_type MotionType
	multiplier  int
	int_data    int
}

type Action struct {
	multiplier  int
	motion      Motion
	verb        Verb
	end_in_mode Mode
	yank        bool
}

type Vim struct {
	mode    Mode
	command []rune
	keys    []VimKey
}

func (vim *Vim) init() {
	vim.keys = append(vim.keys, VimKey{key: 'h', function: motionLeft})
	vim.keys = append(vim.keys, VimKey{key: 'l', function: motionRight})
	vim.keys = append(vim.keys, VimKey{key: 'j', function: motionDown})
	vim.keys = append(vim.keys, VimKey{key: 'k', function: motionUp})
}

func (vim *Vim) ParseAction(key rune) (state ParseActionState, action Action) {
	state = PARSE_ACTION_STATE_INVALID

	// parse the previous commands
	for _, command_key := range vim.command {
		for _, vim_key := range vim.keys {
			if vim_key.key == command_key {
				state = vim_key.function(&action)
				break
			}
		}
	}

	// parse the current command
	for _, vim_key := range vim.keys {
		if vim_key.key == key {
			state = vim_key.function(&action)
			break
		}
	}

	// based on the state, append the key or clear the command
	switch state {
	default:
	case PARSE_ACTION_STATE_IN_PROGRESS:
		vim.command = append(vim.command, key)
	case PARSE_ACTION_STATE_COMPLETE:
		vim.command = []rune{}
	}

	return state, action
}

func (vim *Vim) Perform(action Action, buffer Buffer, cursor Point) (bool, Point) {
	switch action.motion.motion_type {
	default:
	case MOTION_TYPE_LEFT:
		cursor = MoveCursor(buffer, cursor, Point{-1, 0})
		return true, cursor
	case MOTION_TYPE_RIGHT:
		cursor = MoveCursor(buffer, cursor, Point{1, 0})
		return true, cursor
	case MOTION_TYPE_UP:
		cursor = MoveCursor(buffer, cursor, Point{0, -1})
		return true, cursor
	case MOTION_TYPE_DOWN:
		cursor = MoveCursor(buffer, cursor, Point{0, 1})
		return true, cursor
	}

	return false, Point{0, 0}
}

func motionLeft(action *Action) ParseActionState {
	action.motion.motion_type = MOTION_TYPE_LEFT
	return PARSE_ACTION_STATE_COMPLETE
}

func motionRight(action *Action) ParseActionState {
	action.motion.motion_type = MOTION_TYPE_RIGHT
	return PARSE_ACTION_STATE_COMPLETE
}

func motionUp(action *Action) ParseActionState {
	action.motion.motion_type = MOTION_TYPE_UP
	return PARSE_ACTION_STATE_COMPLETE
}

func motionDown(action *Action) ParseActionState {
	action.motion.motion_type = MOTION_TYPE_DOWN
	return PARSE_ACTION_STATE_COMPLETE
}
