package whatnot

import (
	"github.com/stretchr/testify/assert"
	"math/rand"
	"path"
	"testing"
)

var created []string

func TestRandomFiles(t *testing.T) {
	gns := createTestNamespace(t)
	err := gns.GenerateRandomPaths("/", 0)
	assert.Nil(t, err, "randomized pathelement creation failed")
	var i int
	var p string
	for i, p = range created {
		elem := gns.FetchAbsolutePath(PathString(p))
		if !assert.NotNil(t, elem, "created element was not returned") {
			t.Errorf("lost registered path element: %s", p)
			return
		}
	}
	t.Logf("tested %d random path elements successfully", i)
}

var SimpleElementNames = []rune("abcdefghijklmnopqrstuvwxyz01234567890-_")
var ComplexElementNames = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ01234567890!@#$%^&*()-_+= ;.,<>'\"[]{}() ")

var RandomLayout = true

var maxTestElementNameSize = 16
var maxTestPathDepth = 10
var maxTestPathWidth = 20
var maxTestPathEnds = 20
func (ns *Namespace) GenerateRandomPaths(root string, startdepth int) (err error) {
	numfiles := maxTestPathEnds
	if RandomLayout {
		numfiles = rand.Intn(numfiles) + 1
	}

	for i := 0; i < numfiles; i++ {
		if err = ns.createRandomPathElement(root); err != nil {
			return err
		}

	}

	if startdepth+1 <= maxTestPathDepth {
		numdirs := maxTestPathWidth
		if RandomLayout {
			numdirs = rand.Intn(numdirs) + 1
		}

		for i := 0; i < numdirs; i++ {
			if err = ns.randomAbsolutePath(root, startdepth+1) ; err != nil {
				return err
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
	return ns.RegisterAbsolutePath(PathString(filepath).ToAbsolutePath())
}