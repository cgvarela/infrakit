package discovery

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
	"time"

	rpc "github.com/docker/infrakit/rpc/instance"
	"github.com/docker/infrakit/rpc/server"
	"github.com/stretchr/testify/require"
)

func blockWhileFileExists(name string) {
	for {
		_, err := os.Stat(name)
		if err != nil {
			if os.IsNotExist(err) {
				break
			}
		}
		time.Sleep(10 * time.Millisecond)
	}
}

func TestDirDiscovery(t *testing.T) {

	dir, err := ioutil.TempDir("", "infrakit_dir_test")
	require.NoError(t, err)

	name1 := "server1"
	path1 := filepath.Join(dir, name1)
	server1, err := server.StartPluginAtPath(path1, rpc.PluginServer(nil))
	require.NoError(t, err)
	require.NotNil(t, server1)

	name2 := "server2"
	path2 := filepath.Join(dir, name2)
	server2, err := server.StartPluginAtPath(path2, rpc.PluginServer(nil))
	require.NoError(t, err)
	require.NotNil(t, server2)

	discover, err := newDirPluginDiscovery(dir)
	require.NoError(t, err)

	p, err := discover.Find(name1)
	require.NoError(t, err)
	require.NotNil(t, p)

	p, err = discover.Find(name2)
	require.NoError(t, err)
	require.NotNil(t, p)

	// Now we stop the servers
	server1.Stop()
	blockWhileFileExists(path1)

	p, err = discover.Find(name1)
	require.Error(t, err)

	p, err = discover.Find(name2)
	require.NoError(t, err)
	require.NotNil(t, p)

	server2.Stop()

	blockWhileFileExists(path2)

	p, err = discover.Find(name1)
	require.Error(t, err)

	p, err = discover.Find(name2)
	require.Error(t, err)

	list, err := discover.List()
	require.NoError(t, err)
	require.Equal(t, 0, len(list))
}
