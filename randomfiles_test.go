package whatnot

import (
	"github.com/stretchr/testify/assert"
	"math/rand"
	"path"
	"testing"
)

func TestRandomFiles(t *testing.T) {
	gns := createTestNamespace(t)
	err := gns.GenerateRandomPaths("/", 0)
	assert.Nil(t, err, "randomized pathelement creation failed")
}

var SimpleElementNames = []rune("abcdefghijklmnopqrstuvwxyz01234567890-_")
var ComplexElementNames = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ01234567890!@#$%^&*()-_+= ;.,<>'\"[]{}() ")

var RandomLayout = true

var MaxElementNameSize = 16
var MaxPathDepth = 3
var MaxPathWidth = 20
var MaxPathEnds = 5

func (ns *Namespace) GenerateRandomPaths(root string, startdepth int) (err error) {
	numfiles := MaxPathEnds
	if RandomLayout {
		numfiles = rand.Intn(numfiles) + 1
	}

	for i := 0; i < numfiles; i++ {
		if err = ns.createRandomPathElement(root); err != nil {
			return err
		}

	}

	if startdepth+1 <= MaxPathDepth {
		numdirs := MaxPathWidth
		if RandomLayout {
			numdirs = rand.Intn(numdirs) + 1
		}

		for i := 0; i < numdirs; i++ {
			ns.randomAbsolutePath(root, startdepth+1)
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

func (ns *Namespace) createRandomPathElement(root string) error {
	name := ns.randomPathName(rand.Intn(8), SimpleElementNames)
	filepath := path.Join(root, name)
	return ns.RegisterAbsolutePath(PathString(filepath).ToAbsolutePath())
}

func (ns *Namespace) randomAbsolutePath(root string, depth int) {
	if depth > MaxPathDepth {
		return
	}

	n := rand.Intn(MaxElementNameSize-4) + 4
	name := ns.randomPathName(n, SimpleElementNames)
	root = path.Join(root, name)
	ns.GenerateRandomPaths(root, depth)
	return
}
