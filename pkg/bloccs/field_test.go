package bloccs

import (
	"bloccs-server/pkg/event"
	"fmt"
	"strconv"
	"strings"
	"testing"
)

var B = uint8(Bedrock)
var T = tPiece.Name

func printField(data PieceData, width int, height int) {
	fmt.Print("┌───┬" + strings.Repeat("─", width*3+6) + "─┐\n")

	for y := 0; y < height; y++ {
		fmt.Printf("│ %d │", y)

		for x := 0; x < width; x++ {
			i := y*6 + x
			c := data[i]
			s := ""

			if c == 0 {
				s = "."
			} else {
				s = string(c)
			}

			if c < 33 || c > 126 {
				s = strconv.Itoa(int(c))
			}

			fmt.Printf(" %3s", s)
		}

		fmt.Print(" │\n")
	}

	fmt.Print("│   └" + strings.Repeat("─", width*3+7) + "┤\n│    ")

	for x := 0; x < width; x++ {
		fmt.Printf(" %3d", x)
	}

	fmt.Print(" │\n└──" + strings.Repeat("─", width*3+8) + "─┘\n")
}

func assertFieldsEqual(t *testing.T, field *Field, expectedField PieceData) {
	if field.width*field.height > len(expectedField) {
		t.Fatal("field size mismatch")
	}

	for x := 0; x < field.width; x++ {
		for y := 0; y < field.height; y++ {
			i := y*field.width + x

			if field.Data[i] != expectedField[i] {
				fmt.Println("field:")
				printField(field.Data, field.width, field.height)

				fmt.Println("expected field:")
				printField(expectedField, field.width, field.height)

				t.Fatalf("fields not equal at %d;%d", x, y)
			}
		}
	}
}

func TestField_GetId(t *testing.T) {
	field := NewField(nil, "test", 0, 0)
	id := field.GetId()

	if id != "test" {
		t.Fatalf("expected id `test`, got %s", id)
	}
}

func TestField_RLock(t *testing.T) {
	field := NewField(nil, "test", 0, 0)

	field.RLock()

	if field.mu.TryLock() {
		t.Fatal("expected field mutex to be not lockable")
	}
}

func TestField_RUnlock(t *testing.T) {
	field := NewField(nil, "test", 0, 0)

	field.RLock()
	field.RUnlock()

	if !field.mu.TryLock() {
		t.Fatal("expected field mutex to be lockable")
	}
}

func TestField_CanPutPiece(t *testing.T) {
	fieldWidth := 6
	fieldHeight := 6

	field := NewField(nil, "test", fieldWidth, fieldHeight)

	if !field.CanPutPiece(&tPiece, 0, 1, 1) {
		t.Fatal("expected piece be placeable at 1;1")
	}

	if !field.CanPutPiece(&tPiece, 1, 1, 1) {
		t.Fatal("expected piece be placeable at 1;1 with rotation 1")
	}

	if field.CanPutPiece(&tPiece, 0, -1, -1) {
		t.Fatal("expected piece not placeable at -1;-1")
	}

	if field.CanPutPiece(&tPiece, 0, 7, 7) {
		t.Fatal("expected piece not placeable at 7;7")
	}

	if field.CanPutPiece(&tPiece, 1, -1, -1) {
		t.Fatal("expected piece not placeable at -1;-1 with rotation 1")
	}
}

func TestField_IncreaseBedrock(t *testing.T) {
	fieldWidth := 6
	fieldHeight := 6

	field := NewField(nil, "test", fieldWidth, fieldHeight)

	field.IncreaseBedrock(2)

	expectedField := PieceData{
		0, 0, 0, 0, 0, 0,
		0, 0, 0, 0, 0, 0,
		0, 0, 0, 0, 0, 0,
		0, 0, 0, 0, 0, 0,
		B, B, B, B, B, B,
		B, B, B, B, B, B,
	}

	assertFieldsEqual(t, field, expectedField)

	field.IncreaseBedrock(1)

	expectedField = PieceData{
		0, 0, 0, 0, 0, 0,
		0, 0, 0, 0, 0, 0,
		0, 0, 0, 0, 0, 0,
		B, B, B, B, B, B,
		B, B, B, B, B, B,
		B, B, B, B, B, B,
	}

	assertFieldsEqual(t, field, expectedField)

	field.PutPiece(&tPiece, 0, 2, 0)
	field.IncreaseBedrock(1)

	expectedField = PieceData{
		0, 0, T, T, T, 0,
		0, 0, 0, T, 0, 0,
		B, B, B, B, B, B,
		B, B, B, B, B, B,
		B, B, B, B, B, B,
		B, B, B, B, B, B,
	}

	assertFieldsEqual(t, field, expectedField)
}

func TestField_DecreaseBedrock(t *testing.T) {
	fieldWidth := 6
	fieldHeight := 6

	field := NewField(nil, "test", fieldWidth, fieldHeight)

	initialField := PieceData{
		0, 0, T, T, T, 0,
		0, 0, 0, T, 0, 0,
		B, B, B, B, B, B,
		B, B, B, B, B, B,
		B, B, B, B, B, B,
		B, B, B, B, B, B,
	}

	field.setData(initialField, 4)

	producedBedrock := field.DecreaseBedrock(3)

	if producedBedrock != 0 {
		t.Fatalf("expected 0 produced bedrock, got %d", producedBedrock)
	}

	expectedField := PieceData{
		0, 0, 0, 0, 0, 0,
		0, 0, 0, 0, 0, 0,
		0, 0, 0, 0, 0, 0,
		0, 0, T, T, T, 0,
		0, 0, 0, T, 0, 0,
		B, B, B, B, B, B,
	}

	assertFieldsEqual(t, field, expectedField)

	producedBedrock = field.DecreaseBedrock(1)

	if producedBedrock != 0 {
		t.Fatalf("expected 0 produced bedrock, got %d", producedBedrock)
	}

	expectedField = PieceData{
		0, 0, 0, 0, 0, 0,
		0, 0, 0, 0, 0, 0,
		0, 0, 0, 0, 0, 0,
		0, 0, 0, 0, 0, 0,
		0, 0, T, T, T, 0,
		0, 0, 0, T, 0, 0,
	}

	assertFieldsEqual(t, field, expectedField)

	producedBedrock = field.DecreaseBedrock(2)

	if producedBedrock != 2 {
		t.Fatalf("expected 2 produced bedrock, got %d", producedBedrock)
	}

	expectedField = PieceData{
		0, 0, 0, 0, 0, 0,
		0, 0, 0, 0, 0, 0,
		0, 0, 0, 0, 0, 0,
		0, 0, 0, 0, 0, 0,
		0, 0, T, T, T, 0,
		0, 0, 0, T, 0, 0,
	}

	assertFieldsEqual(t, field, expectedField)
}

func TestField_DecreaseBedrockWithProducing(t *testing.T) {
	fieldWidth := 6
	fieldHeight := 6

	field := NewField(nil, "test", fieldWidth, fieldHeight)

	initialField := PieceData{
		0, 0, T, T, T, 0,
		0, 0, 0, T, 0, 0,
		B, B, B, B, B, B,
		B, B, B, B, B, B,
		B, B, B, B, B, B,
		B, B, B, B, B, B,
	}

	field.setData(initialField, 4)

	producedBedrock := field.DecreaseBedrock(7)

	if producedBedrock != 3 {
		t.Fatalf("expected 3 produced bedrock, got %d", producedBedrock)
	}

	expectedField := PieceData{
		0, 0, 0, 0, 0, 0,
		0, 0, 0, 0, 0, 0,
		0, 0, 0, 0, 0, 0,
		0, 0, 0, 0, 0, 0,
		0, 0, T, T, T, 0,
		0, 0, 0, T, 0, 0,
	}

	assertFieldsEqual(t, field, expectedField)
}

func TestField_DecreaseBedrockMoreThanHeight(t *testing.T) {
	fieldWidth := 6
	fieldHeight := 6

	field := NewField(nil, "test", fieldWidth, fieldHeight)

	initialField := PieceData{
		0, 0, T, T, T, 0,
		0, 0, 0, T, 0, 0,
		B, B, B, B, B, B,
		B, B, B, B, B, B,
		B, B, B, B, B, B,
		B, B, B, B, B, B,
	}

	field.setData(initialField, 4)
	gameOver := field.IncreaseBedrock(3)

	if !gameOver {
		t.Fatal("expected game over")
	}

	expectedField := PieceData{
		0, 0, T, T, T, 0,
		0, 0, 0, T, 0, 0,
		B, B, B, B, B, B,
		B, B, B, B, B, B,
		B, B, B, B, B, B,
		B, B, B, B, B, B,
	}

	assertFieldsEqual(t, field, expectedField)
}

func TestField_PutPiece(t *testing.T) {
	fieldWidth := 6
	fieldHeight := 6

	field := NewField(nil, "test", fieldWidth, fieldHeight)

	field.PutPiece(&tPiece, 0, 2, 2)

	expectedField := PieceData{
		0, 0, 0, 0, 0, 0,
		0, 0, 0, 0, 0, 0,
		0, 0, 0, 0, 0, 0,
		0, 0, T, T, T, 0,
		0, 0, 0, T, 0, 0,
		0, 0, 0, 0, 0, 0,
	}

	assertFieldsEqual(t, field, expectedField)
}

func TestField_PutPieceEmitsEvent(t *testing.T) {
	bus := event.NewBus()

	bus.Start()

	eventCount := 0
	expectedEventCount := 1

	field := NewField(bus, "test", 6, 6)

	bus.Subscribe(EventFieldUpdate, func(event *event.Event) {
		if f, _ := event.Source.(*Field); f != field {
			t.Fatalf("event source is not field")
		}

		eventCount += 1
	})

	field.PutPiece(&tPiece, 0, 2, 2)

	bus.Stop()

	if eventCount != expectedEventCount {
		t.Fatalf("missing events. exepected %d, got %d", expectedEventCount, eventCount)
	}
}

func TestField_PutPieceAtBottom(t *testing.T) {
	fieldWidth := 6
	fieldHeight := 6

	field := NewField(nil, "test", fieldWidth, fieldHeight)

	field.PutPiece(&tPiece, 0, 2, 3)

	expectedField := PieceData{
		0, 0, 0, 0, 0, 0,
		0, 0, 0, 0, 0, 0,
		0, 0, 0, 0, 0, 0,
		0, 0, 0, 0, 0, 0,
		0, 0, T, T, T, 0,
		0, 0, 0, T, 0, 0,
	}

	assertFieldsEqual(t, field, expectedField)
}

func TestField_ClearFullRows(t *testing.T) {
	fieldWidth := 6
	fieldHeight := 6

	field := NewField(nil, "test", fieldWidth, fieldHeight)

	initialField := PieceData{
		0, 0, 0, 0, T, 0,
		T, T, T, T, T, T,
		T, T, 0, T, T, T,
		T, T, T, T, T, T,
		B, B, B, B, B, B,
		B, B, B, B, B, B,
	}

	field.setData(initialField, 2)

	clearedRows := field.ClearFullRows()

	expectedField := PieceData{
		0, 0, 0, 0, 0, 0,
		0, 0, 0, 0, 0, 0,
		0, 0, 0, 0, T, 0,
		T, T, 0, T, T, T,
		B, B, B, B, B, B,
		B, B, B, B, B, B,
	}

	if clearedRows != 2 {
		t.Fatalf("expected 2 cleared rows, got %d", clearedRows)
	}

	assertFieldsEqual(t, field, expectedField)
}

func TestField_Reset(t *testing.T) {
	fieldWidth := 6
	fieldHeight := 6

	field := NewField(nil, "test", fieldWidth, fieldHeight)

	initialField := PieceData{
		0, 0, 0, 0, T, 0,
		T, T, T, T, T, T,
		T, T, 0, T, T, T,
		T, T, T, T, T, T,
		B, B, B, B, B, B,
		B, B, B, B, B, B,
	}

	field.setData(initialField, 2)

	field.Reset()

	expectedField := PieceData{
		0, 0, 0, 0, 0, 0,
		0, 0, 0, 0, 0, 0,
		0, 0, 0, 0, 0, 0,
		0, 0, 0, 0, 0, 0,
		0, 0, 0, 0, 0, 0,
		0, 0, 0, 0, 0, 0,
	}

	assertFieldsEqual(t, field, expectedField)

	if field.currentBedrock != 0 {
		t.Fatalf("expected 0 bedrock, got %d", field.currentBedrock)
	}
}
