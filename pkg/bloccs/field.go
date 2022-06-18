package bloccs

import (
	"bloccs-server/pkg/event"
	"sync"
)

const EventFieldUpdate = "field_update"

type Field struct {
	GameId         string    `json:"gameId"`
	Data           PieceData `json:"data"`
	width          int
	height         int
	centerX        int
	currentBedrock int
	eventBus       *event.Bus
	mu             *sync.RWMutex
}

func NewField(bus *event.Bus, id string, w int, h int) *Field {
	f := Field{
		GameId:         id,
		Data:           make(PieceData, w*h),
		width:          w,
		height:         h,
		centerX:        w/2 - pieceDataWidth/2,
		currentBedrock: 0,
		eventBus:       bus,
		mu:             &sync.RWMutex{},
	}

	return &f
}

func (f *Field) RLock() {
	f.mu.RLock()
}

func (f *Field) RUnlock() {
	f.mu.RUnlock()
}

func (f *Field) GetId() string {
	f.mu.RLock()
	f.mu.RUnlock()

	return f.GameId
}

func (f *Field) publishUpdate() {
	if f.eventBus == nil {
		return
	}

	// todo only send if necessary
	f.eventBus.Publish(event.New(EventFieldUpdate, f, nil))
}

func (f *Field) Reset() {
	defer f.mu.Unlock()
	f.mu.Lock()

	f.Data = make(PieceData, f.width*f.height)

	f.currentBedrock = 0

	f.publishUpdate()
}

func putPieceToSlice(buf PieceData, bufWidth int, p *Piece, r int, x int, y int) {
	for px := 0; px < pieceDataWidth; px++ {
		for py := 0; py < pieceDataWidth; py++ {
			d := p.GetDataXY(r, px, py)

			if d != 0 {
				id := (py+y)*bufWidth + (px + x)

				buf[id] = d
			}
		}
	}
}

func (f *Field) PutPiece(p *Piece, r int, x int, y int) {
	defer f.mu.Unlock()
	f.mu.Lock()

	putPieceToSlice(f.Data, f.width, p, r, x, y)

	f.publishUpdate()
}

func (f *Field) isInBounds(x int, y int) bool {
	if x < 0 || x >= f.width || y < 0 || y >= f.height {
		return false
	}

	return true
}

func (f *Field) CanPutPiece(p *Piece, r int, x int, y int) bool {
	defer f.mu.RUnlock()
	f.mu.RLock()

	for px := 0; px < pieceDataWidth; px++ {
		for py := 0; py < pieceDataWidth; py++ {
			tx := px + x
			ty := py + y

			if p.GetDataXY(r, px, py) != 0 && (!f.isInBounds(tx, ty) || f.Data[ty*f.width+tx] != 0) {
				return false
			}
		}
	}

	return true
}

// todo: broken
// applyBedrock returns true if there is a bedrock induced game over
func (f *Field) applyBedrock(bedrockDelta int, producedBedrock *int) bool {
	if bedrockDelta < 0 && f.currentBedrock >= 0 {
		for bedrockDelta < 0 && f.currentBedrock > 0 {
			for y := f.height - 1; y >= 0; y-- {
				for x := 0; x < f.width; x++ {
					blockData := f.getDataXY(x, y)

					if y != f.height-1 {
						f.setDataXY(x, y+1, blockData)
						f.setDataXY(x, y, 0)
					}
				}
			}

			f.currentBedrock--
			bedrockDelta++

			f.publishUpdate()
		}

		if producedBedrock != nil {
			*producedBedrock = bedrockDelta * -1
		}
	} else if bedrockDelta > 0 {
		for bedrockDelta > 0 {
			for y := 0; y < f.height; y++ {
				for x := 0; x < f.width; x++ {
					blockData := f.getDataXY(x, y)

					// todo: check if bedrock crashes into current falling piece, return true

					if blockData != 0 && y == 0 {
						return true
					}

					if y != 0 {
						f.setDataXY(x, y-1, blockData)
					}
				}
			}

			f.currentBedrock++
			bedrockDelta--

			for x := 0; x < f.width; x++ {
				f.setDataXY(x, f.height-f.currentBedrock, Bedrock)
			}

			f.publishUpdate()
		}
	}

	return false
}

// IncreaseBedrock returns true on gameover
func (f *Field) IncreaseBedrock(increaseCount int) bool {
	defer f.mu.Unlock()
	f.mu.Lock()

	if increaseCount+f.currentBedrock > f.height {
		increaseCount = f.height
	}

	return f.applyBedrock(increaseCount, nil)
}

func (f *Field) DecreaseBedrock(decreaseCount int) int {
	defer f.mu.Unlock()
	f.mu.Lock()

	producedBedrock := 0

	f.applyBedrock(-decreaseCount, &producedBedrock)

	return producedBedrock
}

func (f *Field) ClearFullRows() int {
	defer f.mu.Unlock()
	f.mu.Lock()

	cleared := 0

	for y := 0; y < f.height; y++ {
		isFull := true

		for x := 0; x < f.width; x++ {
			d := f.getDataXY(x, y)

			if d == 0 || d == Bedrock {
				isFull = false
				break
			}
		}

		if isFull {
			for yi := y; yi > -1; yi-- {
				for x := 0; x < f.width; x++ {
					if yi == 0 {
						f.setDataXY(x, yi, 0)
					} else {
						f.setDataXY(x, yi, f.getDataXY(x, yi-1))
					}
				}
			}

			cleared++
			y--
		}
	}

	if cleared > 0 {
		f.publishUpdate()
	}

	return cleared
}

func (f *Field) setData(d PieceData, bedrock int) {
	for x := 0; x < f.width; x++ {
		for y := 0; y < f.height; y++ {
			f.setDataXY(x, y, d[x+y*f.width])
		}
	}

	f.currentBedrock = bedrock
}

func (f *Field) setDataXY(x int, y int, d uint8) {
	f.Data[y*f.width+x] = d
}

func (f *Field) getDataXY(x int, y int) uint8 {
	return f.Data[y*f.width+x]
}
