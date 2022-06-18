package bloccs

import (
	"fmt"
	"strings"
)

const pieceDataWidth = 4

const Bedrock = 'B'

type PieceData []uint8

type Piece struct {
	Name         uint8 `json:"name"`
	rotatedFaces *[]PieceData
}

func (p *Piece) RLock() {
}

func (p *Piece) RUnlock() {
}

func (d PieceData) MarshalJSON() ([]byte, error) {
	var result string

	if d == nil {
		result = "null"
	} else {
		result = strings.Join(strings.Fields(fmt.Sprintf("%d", d)), ",")
	}

	return []byte(result), nil
}

func (p *Piece) ClampRotation(rot int) int {
	return rot % len(*p.rotatedFaces)
}

func (p *Piece) GetData(rot int) *PieceData {
	return &((*p.rotatedFaces)[p.ClampRotation(rot)])
}

func (p *Piece) GetDataXY(rot int, x int, y int) uint8 {
	return (*p.rotatedFaces)[p.ClampRotation(rot)][y*pieceDataWidth+x]
}

var iPiece = Piece{
	Name: 'I',
	rotatedFaces: &[]PieceData{
		{
			0, 0, 0, 0,
			0, 0, 0, 0,
			'I', 'I', 'I', 'I',
			0, 0, 0, 0,
		},
		{
			0, 0, 'I', 0,
			0, 0, 'I', 0,
			0, 0, 'I', 0,
			0, 0, 'I', 0,
		},
	},
}

var oPiece = Piece{
	Name: 'O',
	rotatedFaces: &[]PieceData{
		{
			0, 0, 0, 0,
			0, 'O', 'O', 0,
			0, 'O', 'O', 0,
			0, 0, 0, 0,
		},
	},
}

var jPiece = Piece{
	Name: 'J',
	rotatedFaces: &[]PieceData{
		{
			0, 0, 0, 0,
			'J', 'J', 'J', 0,
			0, 0, 'J', 0,
			0, 0, 0, 0,
		},
		{
			0, 'J', 0, 0,
			0, 'J', 0, 0,
			'J', 'J', 0, 0,
			0, 0, 0, 0,
		},
		{
			'J', 0, 0, 0,
			'J', 'J', 'J', 0,
			0, 0, 0, 0,
			0, 0, 0, 0,
		},
		{
			0, 'J', 'J', 0,
			0, 'J', 0, 0,
			0, 'J', 0, 0,
			0, 0, 0, 0,
		},
	},
}

var lPiece = Piece{
	Name: 'L',
	rotatedFaces: &[]PieceData{
		{
			0, 0, 0, 0,
			'L', 'L', 'L', 0,
			'L', 0, 0, 0,
			0, 0, 0, 0,
		},
		{
			'L', 'L', 0, 0,
			0, 'L', 0, 0,
			0, 'L', 0, 0,
			0, 0, 0, 0,
		},
		{
			0, 0, 'L', 0,
			'L', 'L', 'L', 0,
			0, 0, 0, 0,
			0, 0, 0, 0,
		},
		{
			0, 'L', 0, 0,
			0, 'L', 0, 0,
			0, 'L', 'L', 0,
			0, 0, 0, 0,
		},
	},
}

var sPiece = Piece{
	Name: 'S',
	rotatedFaces: &[]PieceData{
		{
			0, 0, 0, 0,
			0, 'S', 'S', 0,
			'S', 'S', 0, 0,
			0, 0, 0, 0,
		},
		{
			0, 'S', 0, 0,
			0, 'S', 'S', 0,
			0, 0, 'S', 0,
			0, 0, 0, 0,
		},
		{
			0, 0, 0, 0,
			0, 'S', 'S', 0,
			'S', 'S', 0, 0,
			0, 0, 0, 0,
		},
		{
			0, 'S', 0, 0,
			0, 'S', 'S', 0,
			0, 0, 'S', 0,
			0, 0, 0, 0,
		},
	},
}

var tPiece = Piece{
	Name: 'T',
	rotatedFaces: &[]PieceData{
		{
			0, 0, 0, 0,
			'T', 'T', 'T', 0,
			0, 'T', 0, 0,
			0, 0, 0, 0,
		},
		{
			0, 'T', 0, 0,
			'T', 'T', 0, 0,
			0, 'T', 0, 0,
			0, 0, 0, 0,
		},
		{
			0, 'T', 0, 0,
			'T', 'T', 'T', 0,
			0, 0, 0, 0,
			0, 0, 0, 0,
		},
		{
			0, 'T', 0, 0,
			0, 'T', 'T', 0,
			0, 'T', 0, 0,
			0, 0, 0, 0,
		},
	},
}

var zPiece = Piece{
	Name: 'Z',
	rotatedFaces: &[]PieceData{
		{
			0, 0, 0, 0,
			'Z', 'Z', 0, 0,
			0, 'Z', 'Z', 0,
			0, 0, 0, 0,
		},
		{
			0, 0, 'Z', 0,
			0, 'Z', 'Z', 0,
			0, 'Z', 0, 0,
			0, 0, 0, 0,
		},
		{
			0, 0, 0, 0,
			'Z', 'Z', 0, 0,
			0, 'Z', 'Z', 0,
			0, 0, 0, 0,
		},
		{
			0, 0, 'Z', 0,
			0, 'Z', 'Z', 0,
			0, 'Z', 0, 0,
			0, 0, 0, 0,
		},
	},
}

var AllPieces = []*Piece{
	&tPiece,
	&lPiece,
	&iPiece,
	&jPiece,
	&oPiece,
	&sPiece,
	&zPiece,
}
