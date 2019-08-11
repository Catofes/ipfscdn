package node

import (
	"context"
	"io"
	"net/http"
	"net/url"
	"time"

	shell "github.com/ipfs/go-ipfs-api"
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
)

type node struct {
	config
	web      *echo.Echo
	ipfs     *shell.Shell
	commands *queue
	peers    map[string]peer
}

func (s *node) init() *node {
	s.web = echo.New()
	s.ipfs = shell.NewShell(s.IpfsAddr)
	s.commands = (&queue{}).init(100)
	//s.nodes = make(map[string]peer)
	s.webBind()
	return s
}

func (s *node) loop() {
	s.handleCommands()
	s.web.Start(s.Listen)
}

func (s *node) webBind() {
	s.web.Use(middleware.Recover())
	auth := middleware.KeyAuth(func(k string, c echo.Context) (bool, error) {
		if k == s.config.Key {
			return true, nil
		}
		return false, nil
	})
	s.web.PUT("/pin/:hash", s.pinFile, auth)
	s.web.DELETE("/pin/:hash", s.unPinFile, auth)
	s.web.PUT("/node/:hash", s.addSwarm, auth)
	s.web.GET("/generate_204", s.generate204)
	s.web.GET("/file/:hash", s.getFile)
}

func (s *node) getFile(ctx echo.Context) error {
	fileHash := ctx.Param("hash")
	if fileHash == "" {
		return ctx.NoContent(http.StatusBadRequest)
	}
	r := ctx.Request()
	hh := http.Header{}
	for k, v := range r.Header {
		hh[k] = v
	}
	rr := &http.Request{
		Method: http.MethodGet,
		Header: hh,
		Close:  r.Close,
	}
	var err error
	rr.URL, err = url.Parse(s.IpfsGateway + "/ipfs/" + fileHash)
	if err != nil {
		return ctx.String(http.StatusBadGateway, err.Error())
	}
	resp, err := http.DefaultClient.Do(rr)
	if err != nil {
		return ctx.String(http.StatusBadGateway, err.Error())
	}
	defer resp.Body.Close()
	rh := ctx.Response().Header()
	for k, v := range resp.Header {
		rh[k] = v
	}
	ctx.Response().WriteHeader(resp.StatusCode)
	w := ctx.Response().Writer

	if resp.ContentLength > 0 {
		io.CopyN(w, resp.Body, resp.ContentLength)
	} else if resp.Close {
		io.Copy(w, resp.Body)
	}
	return nil
}

func (s *node) generate204(ctx echo.Context) error {
	return ctx.NoContent(http.StatusNoContent)
}

func (s *node) pinFile(ctx echo.Context) error {
	fileHash := ctx.Param("hash")
	log.Debug("[%s] pin file.", fileHash)
	pinInfos, err := s.ipfs.Pins()
	if err != nil {
		log.Debug("[%s] cache error, %s.", fileHash, err)
		return ctx.String(http.StatusBadGateway, err.Error())
	}
	if fileHash == "" {
		return ctx.NoContent(http.StatusBadRequest)
	}
	if _, ok := pinInfos[fileHash]; ok {
		return ctx.NoContent(http.StatusOK)
	}
	if c := s.commands.get("pin:" + fileHash); c != nil {
		return ctx.NoContent(http.StatusAccepted)
	}
	if c := s.commands.get("unpin:" + fileHash); c != nil {
		s.commands.del("unpin:" + fileHash)
	}
	cc, cf := context.WithCancel(context.Background())
	s.commands.push(&command{
		c:   "pin",
		a:   fileHash,
		ctx: cc,
		cf:  cf,
	})
	return ctx.NoContent(http.StatusAccepted)
}

func (s *node) unPinFile(ctx echo.Context) error {
	fileHash := ctx.Param("hash")
	log.Debug("[%s] unpin file.", fileHash)
	pinInfos, err := s.ipfs.Pins()
	if err != nil {
		log.Debug("[%s] unpin error, %s.", fileHash, err)
		return ctx.String(http.StatusBadGateway, err.Error())
	}
	if fileHash == "" {
		return ctx.NoContent(http.StatusBadRequest)
	}
	if _, ok := pinInfos[fileHash]; !ok {
		return ctx.NoContent(http.StatusOK)
	}
	if c := s.commands.get("unpin:" + fileHash); c != nil {
		return ctx.NoContent(http.StatusAccepted)
	}
	if c := s.commands.get("pin:" + fileHash); c != nil {
		s.commands.del("pin:" + fileHash)
	}
	cc, cf := context.WithCancel(context.Background())
	s.commands.push(&command{
		c:   "unpin",
		a:   fileHash,
		ctx: cc,
		cf:  cf,
	})
	return ctx.NoContent(http.StatusAccepted)
}

func (s *node) addSwarm(ctx echo.Context) error {
	nodeHash := ctx.Param("hash")
	type nodeAddr struct {
		Addr string
	}
	node := nodeAddr{}
	ctx.Bind(&node)
	log.Debug("[%s] add swarm %s.", nodeHash, node.Addr)
	err := s.ipfs.SwarmConnect(context.Background(), node.Addr)
	if err != nil {
		return ctx.String(http.StatusBadGateway, err.Error())
	}
	return ctx.NoContent(http.StatusOK)
}

func (s *node) getPins(ctx echo.Context) error {
	pinInfos, err := s.ipfs.Pins()
	if err != nil {
		log.Error("Get pins error %s.", err)
		return ctx.String(http.StatusBadGateway, err.Error())
	}
	return ctx.JSON(http.StatusOK, pinInfos)
}

func (s *node) handleCommands() {
	t := func() {
		for {
			c := <-s.commands.c
			s.commands.del(c.string())
			s.handleACommand(c)
		}
	}
	for i := 0; i < s.ThreadNum; i++ {
		go t()
	}
	go func() {
		for {
			time.Sleep(10 * time.Second)
			log.Mark(".queueLen", len(s.commands.commands))
		}
	}()
}

func (s *node) handleACommand(c *command) {
	defer c.cf()
	select {
	case <-c.ctx.Done():
		return
	default:
	}
	var err error
	switch c.c {
	case "pin":
		err = s.ipfs.Request("pin/add", c.a).Option("recursive", true).Exec(c.ctx, nil)
	case "unpin":
		err = s.ipfs.Request("pin/rm", c.a).Option("recursive", true).Exec(c.ctx, nil)
	}
	if err != nil {
		log.Debugf("[%s] %s failed: %s.", c.a, c.c, err)
	}
}

func (s *node) sync() error {
	type request struct {
		NodeID   string
		NodeAddr string
		IpfsPath []string
	}
	// rd := request{
	// 	NodeID:   s.NodeID,
	// 	NodeAddr: s.NodeAddr,
	// }
	ipfsInfo, err := s.ipfs.ID()
	if err != nil {
		log.Warning("Get node info failed: %s.", err.Error())
		return err
	}
	for _, v := range ipfsInfo.Addresses {
		log.Debug(v)
	}
	return nil
}
