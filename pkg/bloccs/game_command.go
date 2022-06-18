package bloccs

const CommandLeft = "L"
const CommandRight = "R"
const CommandDown = "D"
const CommandRotate = "X"
const CommandHardLock = "P"
const CommandHold = "H"

func (g *Game) Command(cmd string) {
	defer g.mu.Unlock()
	g.mu.Lock()

	defer g.fallingPiece.mu.Unlock()
	g.fallingPiece.mu.Lock()

	switch cmd {
	case CommandLeft:
		g.moveFallingPiece(0, -1, 0)
		return
	case CommandRight:
		g.moveFallingPiece(0, 1, 0)
		return
	case CommandDown:
		g.moveFallingPiece(0, 0, 1)
		return
	case CommandRotate:
		g.moveFallingPiece(1, 0, 0)
		return
	case CommandHardLock:
		g.hardLockFallingPiece()
		return
	case CommandHold:
		g.holdFallingPiece()
		return
	default:
		return
	}
}
