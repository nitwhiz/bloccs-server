package server

import (
	"context"
	"errors"
	"fmt"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"log"
	"net/http"
	"os"
	"sync"
	"time"
)

const BloccsServerModeVariable = "BLOCCS_MODE"

const BloccsServerModeLive = "live"
const BloccsServerModeDev = "dev"
const BloccsServerModeTest = "test"

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type Server struct {
	ginEngine           *gin.Engine
	rooms               map[string]*Room
	roomsMutex          *sync.Mutex
	waitGroup           *sync.WaitGroup
	roomWatcherStopFunc context.CancelFunc
}

// todo: POST /rooms/<roomId>/reset

func New() *Server {
	s := &Server{
		ginEngine:           nil,
		rooms:               map[string]*Room{},
		roomsMutex:          &sync.Mutex{},
		waitGroup:           &sync.WaitGroup{},
		roomWatcherStopFunc: nil,
	}

	s.augmentGinEngine()

	return s
}

func (s *Server) CreateRoom() *Room {
	room := NewRoom()

	defer s.roomsMutex.Unlock()
	s.roomsMutex.Lock()

	s.rooms[room.GetId()] = room

	return room
}

func (s *Server) GetRoom(id string) *Room {
	defer s.roomsMutex.Unlock()
	s.roomsMutex.Lock()

	r, ok := s.rooms[id]

	if ok {
		return r
	}

	return nil
}

func (s *Server) connect(roomId string, w http.ResponseWriter, r *http.Request) error {
	room := s.GetRoom(roomId)

	conn, err := upgrader.Upgrade(w, r, nil)

	if err != nil {
		return err
	}

	if room == nil {
		if err = conn.WriteControl(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, "room_has_games_running"), time.Now().Add(time.Second)); err != nil {
			return err
		}

		if err = conn.Close(); err != nil {
			return err
		}

		return errors.New("room_not_found")
	}

	//if room.AreGamesRunning() {
	//	log.Println("games running")
	//
	//	if err = conn.WriteControl(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, "room_has_games_running"), time.Now().Add(time.Second)); err != nil {
	//		return err
	//	}
	//
	//	if err = conn.Close(); err != nil {
	//		return err
	//	}
	//
	//	return errors.New("room_has_games_running")
	//}

	return room.Join(conn)
}

func (s *Server) augmentGinEngine() {
	r := gin.New()

	env := os.Getenv(BloccsServerModeVariable)

	if env == "" {
		env = BloccsServerModeDev
	}

	switch env {
	case BloccsServerModeDev:
		gin.SetMode(gin.DebugMode)
		r.Use(gin.Recovery())
		r.Use(gin.Logger())
		break
	case BloccsServerModeTest:
		gin.SetMode(gin.TestMode)
		r.Use(gin.Recovery())
		break
	case BloccsServerModeLive:
		gin.SetMode(gin.ReleaseMode)
		r.Use(gin.Recovery())
		break
	}

	r.Use(cors.Default())

	r.GET("/ping", getPingHandler())

	r.POST("/rooms", postRoomsHandler(s))

	r.POST("/rooms/:roomId/start", postRoomStartHandler(s))

	r.GET("/rooms/:roomId", getRoomsHandler(s))

	r.GET("/rooms/:roomId/socket", getRoomSocketHandler(s))

	s.ginEngine = r
}

func (s *Server) startHTTPServer() error {
	return s.ginEngine.Run("0.0.0.0:7000")
}

func (s *Server) Stop() {
	if s.roomWatcherStopFunc != nil {
		s.roomWatcherStopFunc()
	}

	s.waitGroup.Wait()
}

func (s *Server) roomWatcher() {
	defer s.waitGroup.Done()
	s.waitGroup.Add(1)

	ctx, cancel := context.WithCancel(context.Background())

	s.roomWatcherStopFunc = cancel

	for {
		select {
		case <-ctx.Done():
			return
		case <-time.After(time.Minute):
			s.roomsMutex.Lock()

			runningRooms := map[string]*Room{}
			playerCount := 0

			for rid, r := range s.rooms {
				if r.ShouldStop() {
					go r.Stop()
				} else {
					runningRooms[rid] = r
					playerCount += r.GetPlayerCount()
				}
			}

			log.Println(fmt.Sprintf("ROOMS: %d/%d", len(runningRooms), len(s.rooms)))
			log.Println(fmt.Sprintf("PLAYERS: %d", playerCount))

			s.rooms = runningRooms

			s.roomsMutex.Unlock()
			break
		}
	}
}

func (s *Server) Start() error {
	go s.roomWatcher()

	if err := s.startHTTPServer(); err != nil {
		return err
	}

	return nil
}
