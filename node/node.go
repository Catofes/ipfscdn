package node

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/rpc"
	"net/url"
	"strings"
	"time"

	rnode "github.com/Catofes/ipfscdn/rpc"
	shell "github.com/ipfs/go-ipfs-api"
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
)

type Node struct {
	config
	web      *echo.Echo
	ipfs     *shell.Shell
	commands *queue
	peers    map[string]*peer
	ipfsAddr []string
}

func (s *Node) init() *Node {
	s.web = echo.New()
	s.ipfs = shell.NewShell(s.IpfsAPI)
	s.commands = (&queue{}).init(100)
	s.peers = make(map[string]*peer)
	s.webBind()
	rnode.RegisterNodeService(s)
	return s
}

func (s *Node) rpcloop() {
	listener, err := net.Listen("tcp", s.RPCListen)
	if err != nil {
		log.Fatal("ListenRPC error:", err)
	}
	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Fatal("Accept error:", err)
		}
		go rpc.ServeConn(conn)
	}
}

func (s *Node) loop() {
	s.handleCommands()
	go s.rpcloop()
	s.web.Start(s.Listen)
}

func (s *Node) webBind() {
	s.web.Use(middleware.Recover())
	auth := middleware.KeyAuth(func(k string, c echo.Context) (bool, error) {
		if k == s.config.Key {
			return true, nil
		}
		return false, nil
	})
	s.web.PUT("/pin/:hash", s.webPin, auth)
	s.web.DELETE("/pin/:hash", s.webUnPin, auth)
	s.web.PUT("/node/:hash", s.webConnect, auth)
	s.web.GET("/file/:hash", s.webList, auth)
	s.web.GET("/generate_204", s.generate204)
}

func (s *Node) getFile(ctx echo.Context) error {
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

func (s *Node) generate204(ctx echo.Context) error {
	return ctx.NoContent(http.StatusNoContent)
}

var _ rnode.NodeService = (*Node)(nil)

func (s *Node) Ping(_ bool, result *bool) error {
	return nil
}

func (s *Node) pin(fileHash string) (bool, error) {
	log.Debug("[%s] pin file.", fileHash)
	pinInfos, err := s.ipfs.Pins()
	if err != nil {
		log.Debug("[%s] cache error, %s.", fileHash, err)
		return false, err
	}
	if fileHash == "" {
		return false, errors.New("empty file hash")
	}
	if _, ok := pinInfos[fileHash]; ok {
		return true, nil
	}
	if c := s.commands.get("pin:" + fileHash); c != nil {
		return false, nil
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
	return false, nil
}

func (s *Node) webPin(ctx echo.Context) error {
	fileHash := ctx.Param("hash")
	result, err := s.pin(fileHash)
	if err != nil {
		return ctx.NoContent(http.StatusBadRequest)
	} else if result {
		return ctx.NoContent(http.StatusOK)
	} else {
		return ctx.NoContent(http.StatusAccepted)
	}
}

func (s *Node) Pin(fileHash string, result *bool) error {
	status, err := s.pin(fileHash)
	result = &status
	return err
}

func (s *Node) unPin(fileHash string) (bool, error) {
	log.Debug("[%s] unpin file.", fileHash)
	pinInfos, err := s.ipfs.Pins()
	if err != nil {
		log.Debug("[%s] unpin error, %s.", fileHash, err)
		return false, err
	}
	if fileHash == "" {
		return false, nil
	}
	if _, ok := pinInfos[fileHash]; !ok {
		return true, nil
	}
	if c := s.commands.get("unpin:" + fileHash); c != nil {
		return false, nil
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
	return false, nil
}

func (s *Node) UnPin(fileHash string, result *bool) error {
	status, err := s.unPin(fileHash)
	result = &status
	return err
}

func (s *Node) webUnPin(ctx echo.Context) error {
	fileHash := ctx.Param("hash")
	result, err := s.pin(fileHash)
	if err != nil {
		return ctx.NoContent(http.StatusBadRequest)
	} else if result {
		return ctx.NoContent(http.StatusOK)
	} else {
		return ctx.NoContent(http.StatusAccepted)
	}
}

func (s *Node) connect(node rnode.NodeInfo) error {
	log.Debug("[%s] add swarm %v.", node.NodeID, node.IpfsPath)
	if v, ok := s.peers[node.NodeID]; ok {
		v.addPath(node.IpfsPath, node.NodeAddr)
	} else {
		p := (&peer{NodeID: node.NodeID}).init(s)
		p.addPath(node.IpfsPath, node.NodeAddr)
		s.peers[node.NodeID] = p
		go p.loop()
	}
	return nil
}

func (s *Node) Connect(node rnode.NodeInfo, _ *bool) error {
	return s.connect(node)
}

func (s *Node) webConnect(ctx echo.Context) error {
	node := rnode.NodeInfo{}
	ctx.Bind(&node)
	err := s.connect(node)
	if err != nil {
		return ctx.String(http.StatusBadRequest, err.Error())
	}
	return ctx.NoContent(http.StatusOK)
}

func (s *Node) list() (*rnode.FileLists, error) {
	pinInfos, err := s.ipfs.Pins()
	if err != nil {
		log.Error("Get pins error %s.", err)
		return nil, err
	}
	return &pinInfos, nil
}

func (s *Node) List(_ bool, lists *rnode.FileLists) error {
	lists, err := s.list()
	return err
}

func (s *Node) webList(ctx echo.Context) error {
	pinInfos, err := s.list()
	if err != nil {
		return ctx.String(http.StatusBadGateway, err.Error())
	}
	return ctx.JSON(http.StatusOK, pinInfos)
}

func (s *Node) handleCommands() {
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

func (s *Node) handleACommand(c *command) {
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

func (s *Node) sync() error {
	// rd := request{
	// 	NodeID:   s.NodeID,
	// 	NodeAddr: s.NodeAddr,
	// }
	ipfsInfo, err := s.ipfs.ID()
	if err != nil {
		log.Warning("Get node info failed: %s.", err.Error())
		return err
	}
	s.ipfsAddr = make([]string, 0)

	for _, v := range ipfsInfo.Addresses {
		t := strings.Split(v, "/")
		if len(t) < 2 {
			continue
		}
		addr := net.ParseIP(t[1])
		if addr == nil {
			continue
		}
		if isPrivateIP(addr) {
			continue
		} else {
			s.ipfsAddr = append(s.ipfsAddr, v+ipfsInfo.ID)
		}
	}

	r := nodeInfo{
		NodeID:   ipfsInfo.ID,
		NodeAddr: s.config.NodeAddr,
		IpfsPath: s.ipfsAddr,
		DiskSize: s.config.DiskSize,
		Type:     s.config.Type,
	}
	result := make(map[string]nodeInfo)
	response, err := rest.R().
		SetAuthToken(s.config.Key).
		SetBody(r).
		SetResult(&result).
		Put("%s/init/")
	if err != nil {
		return err
	}
	if response.StatusCode() != http.StatusOK {
		return fmt.Errorf("sync to manager failed, http code: %d", response.StatusCode())
	}
	for _, v := range result {
		if p, ok := s.peers[v.NodeID]; ok {
			p.addPath(v.IpfsPath, v.NodeAddr)
		} else {
			p := (&peer{NodeID: v.NodeID}).init(s)
			p.addPath(v.IpfsPath, v.NodeAddr)
			s.peers[v.NodeID] = p
			go p.loop()
		}
	}

	return nil
}

func (s *Node) syncLoop() {
	for {
		err := s.sync()
		if err != nil {
			log.Debugf("Sync to manager failed: %s.", err.Error())
			time.Sleep(1 * time.Minute)
			continue
		}
		time.Sleep(1 * time.Hour)
	}
}
