package whatnot

import (
	"github.com/stretchr/testify/assert"
	"golang.org/x/sync/errgroup"
	"math/rand"
	"path"
	"testing"
	"time"
)

var manager *NameSpaceManager

func TestNamespaces(t *testing.T) {
	t.Run("Register a Namespace", registerNewNamespace)
	t.Run("Populate Namespace randomly", populateNamespaceWithRandomFiles)
}

func registerNewNamespace(t *testing.T) {
	const testNameSpace = "globaltest"
	t.Log("Creating Namespace Manager")
	manager, err := NewNamespaceManager()
	if !assert.Nil(t, err, "NewNamespaceManager returned error") {
		return
	}
	gns := NewNamespace(testNameSpace)
	err = manager.RegisterNamespace(gns)
	if !assert.Nil(t, err, "RegisterNamespace returned error") {
		return
	}

	fns, err := manager.FetchNamespace(testNameSpace)
	if !assert.Nil(t, err, "FetchNamespace returned error") {
		return
	}
	assert.Equal(t, gns.name, fns.name, "namespace identifiers did not match")

}

var SimpleElementNames = []rune("abcdefghijklmnopqrstuvwxyz01234567890-_")
var ComplexElementNames = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ01234567890!@#$%^&*()-_+= ;.,<>'\"[]{}() ")

// randomize path depth - will massive reduce overall path generation
// and produced a 'jagged' namespace
var randomLayout = false
var linearCreate = false // generate path elements as they are named, or do as concurrent batch
var maxTestElementNameSize = 32
var maxTestPathDepth = 5 // careful changing this one, since its the primary exponent
var maxTestPathWidth = 5

// how many trailing 'files' to randomly generate - this can be quite high since it is not an exponent
var maxTestPathEnds = 7

var created []string

// random generator
var src = rand.NewSource(time.Now().UnixNano())
var r = rand.New(src)

func populateNamespaceWithRandomFiles(t *testing.T) {
	gns := createTestNamespace(t)
	err := gns.GenerateRandomPaths("/", 0)
	assert.Nil(t, err, "randomized pathelement creation failed")
	var i int
	var p string

	if linearCreate == false {
		t.Logf("concurrently creating %d absolute paths", len(created))
		eg := errgroup.Group{}
		for i, p = range created {
			e := p
			eg.Go(func() error {
				return gns.RegisterAbsolutePath(PathString(e).ToAbsolutePath())
			})
		}
		err := eg.Wait()
		if !assert.Nil(t, err, "concurrent creation of path elements did not succeed") {
			t.Error(err.Error())
		}
	}

	t.Logf("check for successful creation of %d absolute paths", len(created))
	for i, p = range created {
		elem := gns.FetchAbsolutePath(PathString(p))
		if !assert.NotNil(t, elem, "created element was not returned") {
			t.Errorf("lost registered path element: %s after %d matches", p, i-1)
			return
		}
	}
	t.Logf("tested %d random path elements successfully", i)
}

func (ns *Namespace) GenerateRandomPaths(root string, startdepth int) (err error) {
	numfiles := maxTestPathEnds
	if randomLayout {
		numfiles = r.Intn(numfiles) + 1
	}

	for i := 0; i < numfiles; i++ {
		if err = ns.createRandomPathElement(root); err != nil {
			return err
		}
	}

	if startdepth+1 <= maxTestPathDepth {
		numdirs := maxTestPathWidth
		if randomLayout {
			numdirs = r.Intn(numdirs) + 1
		}

		for i := 0; i < numdirs; i++ {
			if randomLayout {
				if r.Intn(2) != 0 { // randomize created or not
					if err = ns.randomAbsolutePath(root, startdepth+1); err != nil {
						return err
					}
				}
			} else {
				if err = ns.randomAbsolutePath(root, startdepth+1); err != nil {
					return err
				}
			}
		}
	}

	return err
}

func (ns *Namespace) randomPathName(length int, alphabet []rune) string {
	b := make([]rune, length)
	for i := range b {
		b[i] = alphabet[rand.Intn(len(alphabet))]
	}
	return string(b)
}

func (ns *Namespace) randomAbsolutePath(root string, depth int) (err error) {
	if depth > maxTestPathDepth {
		return
	}

	n := rand.Intn(maxTestElementNameSize-4) + 4
	name := ns.randomPathName(n, SimpleElementNames)
	root = path.Join(root, name)

	return ns.GenerateRandomPaths(root, depth)
}

func (ns *Namespace) createRandomPathElement(root string) error {
	name := ns.randomPathName(rand.Intn(8), SimpleElementNames)
	filepath := path.Join(root, name)
	created = append(created, filepath)
	if linearCreate {
		return ns.RegisterAbsolutePath(PathString(filepath).ToAbsolutePath())
	} else {
		return nil
	}
}
