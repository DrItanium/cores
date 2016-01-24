// iris1 core with unicornhat interface

// +build linux
// +build arm

package unicorn

import (
	"fmt"
	"github.com/DrItanium/cores/iris16"
	"github.com/DrItanium/cores/registration/machine"
	"github.com/DrItanium/cores/registration/parser"
	"github.com/DrItanium/unicornhat"
)

const (
	baseAddress = 0x1000
)

func RegistrationName() string {
	return "iris1-unicorn"
}

func generateCore(a ...interface{}) (machine.Machine, error) {
	return New()
}

func init() {
	machine.Register(RegistrationName(), machine.Registrar(generateCore))
}

func New() (*iris1.Core, error) {
	if c, err := iris1.New(); err != nil {
		return c, err
	} else {
		if err := c.RegisterIoDevice(NewLedArray(baseAddress)); err != nil {
			return nil, err
		} else {
			return c, nil
		}
	}
}

var unicornhatInitialized bool

const (
	commandCell = iota
	brightnessCell
	xCell
	yCell
	redChannel
	blueChannel
	greenChannel
	numCells
)

type ledCommand iris1.Word

const (
	ledCommandClear ledCommand = iota
	ledCommandSetPixel
	ledCommandGetPixel
	ledCommandShow
	ledCommandSetBrightness
	ledCommandGetBrightness
	numLedCommands
)

type LedArray struct {
	base             iris1.Word
	brightness       iris1.Word
	x, y             iris1.Word
	red, green, blue iris1.Word
	initialized      bool
}

func (this *LedArray) Store(address, value iris1.Word) error {
	if address == (this.base + commandCell) {
		cmd := ledCommand(value)
		if cmd >= numLedCommands {
			return fmt.Errorf("Illegal unicornhat opcode %x", cmd)
		} else {
			switch cmd {
			case ledCommandClear:
				unicornhat.ClearLEDBuffer()
			case ledCommandSetPixel:
				if pos, err := unicornhat.PixelPosition(int(this.x), int(this.y)); err != nil {
					return err
				} else {
					unicornhat.SetPixelColor(pos, byte(this.red), byte(this.green), byte(this.blue))
				}
			case ledCommandGetPixel:
				if pos, err := unicornhat.PixelPosition(int(this.x), int(this.y)); err != nil {
					return err
				} else {
					pix := unicornhat.GetPixelColor(pos)
					this.red, this.green, this.blue = iris1.Word(pix.R), iris1.Word(pix.G), iris1.Word(pix.B)
				}
			case ledCommandSetBrightness:
				setBrightness(this.brightness)
			case ledCommandGetBrightness:
				this.brightness = getBrightness()
			case ledCommandShow:
				unicornhat.Show()
			}
			return nil
		}
	} else if address == (this.base + brightnessCell) {
		setBrightness(value)
		return nil
	} else if address == (this.base + xCell) {
		this.x = value
		return nil
	} else if address == (this.base + yCell) {
		this.y = value
		return nil
	} else if address == (this.base + redChannel) {
		this.red = value
		return nil
	} else if address == (this.base + blueChannel) {
		this.blue = value
		return nil
	} else if address == (this.base + greenChannel) {
		this.green = value
		return nil
	} else {
		return fmt.Errorf("Illegal address %x provided!", address)
	}
}

func (this *LedArray) Load(address iris1.Word) (iris1.Word, error) {
	if address == (this.base + commandCell) {
		return 0, fmt.Errorf("Can't read from the command cell of the unicornhat")
	} else if address == (this.base + brightnessCell) {
		return getBrightness(), nil
	} else if address == (this.base + xCell) {
		return this.x, nil
	} else if address == (this.base + yCell) {
		return this.y, nil
	} else if address == (this.base + redChannel) {
		return this.red, nil
	} else if address == (this.base + blueChannel) {
		return this.blue, nil
	} else if address == (this.base + greenChannel) {
		return this.green, nil
	} else {
		return 0, fmt.Errorf("Illegal address %x provided!", address)
	}
}

func setBrightness(brightness iris1.Word) {
	unicornhat.SetBrightness(byte(brightness))
}
func getBrightness() iris1.Word {
	return iris1.Word(unicornhat.GetBrightness())
}
func NewLedArray(baseAddr iris1.Word) *LedArray {
	var l LedArray
	l.base = baseAddr
	return &l
}

func (this *LedArray) Begin() iris1.Word {
	return this.base
}
func (this *LedArray) End() iris1.Word {
	return this.base + numCells
}
func (this *LedArray) RespondsTo(address iris1.Word) bool {
	return this.base <= address && ((this.base + numCells) >= address)
}
func (this *LedArray) Startup() error {
	if this.initialized {
		return fmt.Errorf("Attempted to startup the led array a second time!")
	} else {
		if !unicornhatInitialized {
			unicornhat.SetBrightness(unicornhat.DefaultBrightness / 2)
			if err := unicornhat.Initialize(); err != nil {
				return err
			}
			unicornhat.ClearLEDBuffer()
			unicornhat.Show()
			unicornhatInitialized = true
		}
		this.initialized = true
		return nil
	}
}

func (this *LedArray) Shutdown() error {
	if !this.initialized {
		return fmt.Errorf("Can't shutdown the LedArray when it has either been shutdown or never initialized!")
	} else {
		if unicornhatInitialized {
			unicornhat.Shutdown()
			unicornhatInitialized = false
		}
		this.initialized = false
		return nil
	}
}

func generateParser(args ...interface{}) (parser.Parser, error) {
	// this is a bit of a hack but just call the iris1 parser from the parser list :D
	return parser.New(iris1.RegistrationName(), args)
}

func init() {
	parser.Register(RegistrationName(), parser.Registrar(generateParser))
}
