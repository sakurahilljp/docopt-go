package docopt

import (
	"reflect"
	"strings"
	"testing"
)

func TestOptsUsage(t *testing.T) {
	usage := "Usage: sleep <seconds> [--now]"
	var opts Opts

	opts, _ = ParseArgs(usage, []string{"10"}, "")
	i, err := opts.Int("<seconds>")
	if err != nil || !reflect.DeepEqual(i, int(10)) {
		t.Fail()
	}
	f, err := opts.Float64("<seconds>")
	if err != nil || !reflect.DeepEqual(f, float64(10)) {
		t.Fail()
	}

	opts, _ = ParseArgs(usage, []string{"ten"}, "")
	s, err := opts.String("<seconds>")
	if err != nil || !reflect.DeepEqual(s, string("ten")) {
		t.Fail()
	}

	opts, _ = ParseArgs(usage, []string{"10", "--now"}, "")
	b, err := opts.Bool("--now")
	if err != nil || !reflect.DeepEqual(b, true) {
		t.Fail()
	}
}

func TestOptsErrors(t *testing.T) {
	usage := "Usage: sleep <seconds> [--now]"
	var opts Opts
	var err error

	opts, _ = ParseArgs(usage, []string{"ten!"}, "")

	_, err = opts.Int("<seconds>") // errStrconv
	if err == nil {
		t.Fail()
	}
	_, err = opts.Float64("<seconds>") // errStrconv
	if err == nil {
		t.Fail()
	}

	_, err = opts.Bool("<seconds>") // errType
	if err == nil {
		t.Fail()
	}
	_, err = opts.String("--now") // errType
	if err == nil {
		t.Fail()
	}
	_, err = opts.Int("--now") // errType
	if err == nil {
		t.Fail()
	}
	_, err = opts.Float64("--now") // errType
	if err == nil {
		t.Fail()
	}

	_, err = opts.Int("<missing>") // errKey
	if err == nil {
		t.Fail()
	}
	_, err = opts.Float64("<missing>") // errKey
	if err == nil {
		t.Fail()
	}
	_, err = opts.Bool("<missing>") // errKey
	if err == nil {
		t.Fail()
	}
	_, err = opts.String("<missing>") // errKey
	if err == nil {
		t.Fail()
	}
}

type testOptions struct {
	Command string `docopt:"<command>"`
	Help    bool   `docopt:"-h,--help"`
	Verbose bool   `docopt:"-v"`
	F       bool
}

func TestOptsBind(t *testing.T) {
	var testParser = &Parser{HelpHandler: NoHelpHandler, SkipHelpFlags: true}
	const usage = "Usage: prog [-h|--help] [-v] [-f] <command>"
	for i, c := range []struct {
		argv   []string
		expect testOptions
	}{
		{[]string{"-v", "test_cmd"}, testOptions{
			Command: "test_cmd",
			Help:    false,
			Verbose: true,
			F:       false,
		}},
		{[]string{"-h", "test_cmd"}, testOptions{
			Command: "test_cmd",
			Help:    true,
			Verbose: false,
			F:       false,
		}},
		{[]string{"--help", "test_cmd"}, testOptions{
			Command: "test_cmd",
			Help:    true,
			Verbose: false,
			F:       false,
		}},
		{[]string{"-f", "test_cmd"}, testOptions{
			Command: "test_cmd",
			Help:    false,
			Verbose: false,
			F:       true,
		}},
	} {
		result := testOptions{}
		v, err := testParser.ParseArgs(usage, c.argv, "")
		t.Logf("argv: %v opts: %v", c.argv, v)
		if err != nil {
			t.Fatalf("testcase: %d parse err: %q", i, err)
		}
		if err := v.Bind(&result); err != nil {
			t.Fatalf("testcase: %d bind err: %q", i, err)
		}
		if reflect.DeepEqual(result, c.expect) != true {
			t.Errorf("testcase: %d result: %#v expect: %#v\n", i, result, c.expect)
		}
	}
}

type testTypedOptions struct {
	secret int `docopt:"-s"`

	V       bool
	Number  int16
	Idle    float32
	Pointer uintptr     `docopt:"<ptr>"`
	Ints    []int       `docopt:"<values>"`
	Strings []string    `docopt:"STRINGS"`
	Iface   interface{} `docopt:"IFACE"`
}

func TestBindErrors(t *testing.T) {
	var testParser = &Parser{HelpHandler: NoHelpHandler, SkipHelpFlags: true}
	for i, tc := range []struct {
		usage       string
		command     string
		expectedErr string
	}{
		{
			`Usage: prog [-s]`,
			`prog`,
			`mapping of "-s" is not found in given struct, or is an unexported field`,
		},
		{
			`Usage: prog [--v]`,
			`prog`,
			`mapping of "--v" is not found in given struct, or is an unexported field`,
		},
		{
			`Usage: prog [--number]`,
			`prog`,
			`value of "--number" is not assignable to "Number" field`,
		},
		{
			`Usage: prog [--number=X]`,
			`prog --number=abc`,
			`value of "--number" is not assignable to "Number" field`,
		},
		{
			`Usage: prog <ptr>`,
			`prog 123`,
			`value of "<ptr>" is not assignable to "Pointer" field`,
		},
		{
			`Usage: prog [<values>...]`,
			`prog 123 456`,
			`value of "<values>" is not assignable to "Ints" field`,
		},
		{
			`Usage: prog [-] [IFACE ...]`,
			`prog - 123 456 asd`,
			`mapping of "-" is not found in given struct, or is an unexported field`,
		},
	} {
		argv := strings.Split(tc.command, " ")[1:]
		opts, err := testParser.ParseArgs(tc.usage, argv, "")
		if err != nil {
			t.Fatalf("testcase: %d parse err: %q", i, err)
		}
		var o testTypedOptions
		t.Logf("%#v\n", opts)
		if err := opts.Bind(&o); err != nil {
			if err.Error() != tc.expectedErr {
				t.Fatalf("testcase: %d result: %q expect: %q", i, err.Error(), tc.expectedErr)
			}
		} else {
			t.Fatal("error expected")
		}
	}
}

func TestBindSuccess(t *testing.T) {
	var testParser = &Parser{HelpHandler: NoHelpHandler, SkipHelpFlags: true}
	for i, tc := range []struct {
		usage   string
		command string
	}{
		{
			`Usage: prog [-v]`,
			`prog -v`,
		},
		{
			`Usage: prog [--number=X]`,
			`prog --number=123`,
		},
		{
			`Usage: prog <number>`,
			`prog 123`,
		},
		{
			`Usage: prog NUMBER`,
			`prog 123`,
		},
		{
			`Usage: prog [--idle=X]`,
			`prog --idle=4.1`,
		},
		{
			`Usage: prog [STRINGS ...]`,
			`prog 123 456 asd`,
		},
		{
			`Usage: prog [--help]`,
			`prog --help`,
		},
	} {
		argv := strings.Split(tc.command, " ")[1:]
		opts, err := testParser.ParseArgs(tc.usage, argv, "")
		if err != nil {
			t.Fatalf("testcase: %d parse err: %q", i, err)
		}
		var o testTypedOptions
		t.Logf("%#v\n", opts)
		if err := opts.Bind(&o); err != nil {
			t.Fatalf("testcase: %d error: %q", i, err.Error())
		}
	}
}

func TestBindSimpleStruct(t *testing.T) {
	var testParser = &Parser{HelpHandler: NoHelpHandler, SkipHelpFlags: true}
	opts, err := testParser.ParseArgs("Usage: prog [--number=X]", []string{"--number=123"}, "")
	if err != nil {
		t.Fatal(err)
	}
	var opt struct{ Number int }
	if err := opts.Bind(&opt); err != nil {
		t.Fatal(err)
	}
	if opt.Number != 123 {
		t.Fail()
	}
}

func TestBindToStructWhichAlreadyHasValue(t *testing.T) {
	var testParser = &Parser{HelpHandler: NoHelpHandler, SkipHelpFlags: true}
	opts, err := testParser.ParseArgs("Usage: prog [--number=X]", []string{"--number=123"}, "")
	if err != nil {
		t.Fatal(err)
	}
	var opt = struct{ Number int }{1}
	if err := opts.Bind(&opt); err == nil {
		t.Fatal("error expected")
	}
}

func TestBindDashTag(t *testing.T) {
	var testParser = &Parser{HelpHandler: NoHelpHandler, SkipHelpFlags: true}
	opts, err := testParser.ParseArgs("Usage: prog [-]", []string{"-"}, "")
	if err != nil {
		t.Fatal(err)
	}
	var opt struct {
		Dash bool `docopt:"-"`
	}
	if err := opts.Bind(&opt); err != nil {
		t.Fatal(err)
	}
	if !opt.Dash {
		t.Fail()
	}
}

func TestBindDoubleDashTag(t *testing.T) {
	var testParser = &Parser{HelpHandler: NoHelpHandler, SkipHelpFlags: true}
	opts, err := testParser.ParseArgs("Usage: prog [--]", []string{"--"}, "")
	if err != nil {
		t.Fatal(err)
	}
	var opt struct {
		DoubleDash bool `docopt:"--"`
	}
	if err := opts.Bind(&opt); err != nil {
		t.Fatal(err)
	}
	if !opt.DoubleDash {
		t.Fail()
	}
}

func TestBindHyphenatedTags(t *testing.T) {
	var testParser = &Parser{HelpHandler: NoHelpHandler, SkipHelpFlags: true}
	opts, err := testParser.ParseArgs("Usage: prog --opt-one=N --opt-two=N", []string{"--opt-one", "123", "--opt-two", "234"}, "")
	if err != nil {
		t.Fatal(err)
	}
	var opt struct {
		OptOne string
		OptTwo string
	}
	if err := opts.Bind(&opt); err != nil {
		t.Fatal(err)
	}
	if opt.OptOne != "123" || opt.OptTwo != "234" {
		t.Fail()
	}
}

func TestBindingAnonymousStruct(t *testing.T) {

	type LogOption struct {
		LogLevel string `docopt:"--loglevel"`
	}

	type EmbeddedOption struct {
		LogOption
		Tag   string `docopt:"--tag"`
		Untag string
	}

	var testParser = &Parser{HelpHandler: NoHelpHandler, SkipHelpFlags: true}
	opts, err := testParser.ParseArgs("Usage: prog --tag=TAG --loglevel=LEVEL --untag=UNTAG",
		[]string{"--loglevel", "DEBUG", "--tag", "TAG", "--untag", "UNTAG"}, "")
	if err != nil {
		t.Fatal(err)
	}

	expected := EmbeddedOption{
		LogOption: LogOption{LogLevel: "DEBUG"},
		Tag:       "TAG",
		Untag:     "UNTAG",
	}

	var opt EmbeddedOption

	if err := opts.Bind(&opt); err != nil {
		t.Fatal(err)
	}

	if reflect.DeepEqual(opt, expected) != true {
		t.Errorf("result: %#v expect: %#v\n", opt, expected)
	}
}

func TestBindingMultipleTags(t *testing.T) {

	type MultipleTags struct {
		Multi bool `docopt:"publish,pub"`
	}

	var testParser = &Parser{HelpHandler: NoHelpHandler, SkipHelpFlags: true}

	tags := []string{"pub", "publish"}
	for _, tag := range tags {
		opts, err := testParser.ParseArgs("Usage: prog (publish|pub)",
			[]string{tag}, "")
		if err != nil {
			t.Fatal(tag, err)
		}

		expected := MultipleTags{
			Multi: true,
		}

		var s MultipleTags

		if err := opts.Bind(&s); err != nil {
			t.Fatal(tag, err)
		}

		if reflect.DeepEqual(s, expected) != true {
			t.Errorf("result: %#v expect: %#v\n", s, expected)
		}
	}
}

func TestBindingMultiLevelAnonymousStruct(t *testing.T) {

	type NetworkOptions struct {
		Port int `docopt:"--port"` // Tagged field in innermost struct
	}

	type ServiceOptions struct {
		NetworkOptions        // Embed NetworkOptions
		ServiceName    string // Auto-mapped field in middle struct (--service-name)
	}

	type AppConfig struct {
		ServiceOptions        // Embed ServiceOptions
		Verbose        bool   `docopt:"-v"`       // Tagged field in outermost struct
		ConfigFile     string `docopt:"--config"` // Tagged field in outermost struct
		AppName        string // Auto-mapped field in outermost struct (--app-name)
	}

	var testParser = &Parser{HelpHandler: NoHelpHandler, SkipHelpFlags: true}
	usage := `Usage: myapp [-v] --port=PORT --service-name=NAME --config=FILE --app-name=APPNAME`
	argv := []string{
		"-v",
		"--port", "8080",
		"--service-name", "auth-service",
		"--config", "/etc/myapp.conf",
		"--app-name", "MyApplication",
	}

	opts, err := testParser.ParseArgs(usage, argv, "")
	if err != nil {
		t.Fatalf("ParseArgs failed: %v", err)
	}

	expected := AppConfig{
		ServiceOptions: ServiceOptions{
			NetworkOptions: NetworkOptions{
				Port: 8080,
			},
			ServiceName: "auth-service",
		},
		Verbose:    true,
		ConfigFile: "/etc/myapp.conf",
		AppName:    "MyApplication",
	}

	var actual AppConfig
	if err := opts.Bind(&actual); err != nil {
		t.Fatalf("Bind failed: %v", err)
	}

	if !reflect.DeepEqual(actual, expected) {
		t.Errorf("Bind result mismatch:\n Got: %#v\nWant: %#v", actual, expected)
	}
}

// Test binding to a struct with multiple fields, each having a docopt tag.
func TestBindingMultipleTaggedFields(t *testing.T) {
	type MultiTaggedOptions struct {
		Verbose bool   `docopt:"-v,--verbose"`  // Multiple tags for one field
		Output  string `docopt:"--output-file"` // Tagged string field
		Count   int    `docopt:"<count>"`       // Tagged positional int field
		Mode    string `docopt:"--mode"`        // Tagged string option
	}

	var testParser = &Parser{HelpHandler: NoHelpHandler, SkipHelpFlags: true}
	usage := `Usage: mytool [-v|--verbose] --output-file=FILE --mode=MODE <count>`
	argv := []string{
		"--verbose",                    // Activate the boolean flag
		"--output-file", "results.log", // Provide value for tagged string option
		"--mode", "process", // Provide value for another tagged string option
		"15", // Provide value for tagged positional argument
	}

	opts, err := testParser.ParseArgs(usage, argv, "")
	if err != nil {
		t.Fatalf("ParseArgs failed: %v", err)
	}

	expected := MultiTaggedOptions{
		Verbose: true,
		Output:  "results.log",
		Count:   15,
		Mode:    "process",
	}

	var actual MultiTaggedOptions
	if err := opts.Bind(&actual); err != nil {
		t.Fatalf("Bind failed: %v", err)
	}

	if !reflect.DeepEqual(actual, expected) {
		t.Errorf("Bind result mismatch:\n Got: %#v\nWant: %#v", actual, expected)
	}

	// Test again using the short flag alternative
	argvShort := []string{
		"-v", // Use short flag
		"--output-file", "short.log",
		"--mode", "test",
		"5",
	}
	optsShort, err := testParser.ParseArgs(usage, argvShort, "")
	if err != nil {
		t.Fatalf("ParseArgs (short flag) failed: %v", err)
	}

	expectedShort := MultiTaggedOptions{
		Verbose: true,
		Output:  "short.log",
		Count:   5,
		Mode:    "test",
	}

	var actualShort MultiTaggedOptions
	if err := optsShort.Bind(&actualShort); err != nil {
		t.Fatalf("Bind (short flag) failed: %v", err)
	}

	if !reflect.DeepEqual(actualShort, expectedShort) {
		t.Errorf("Bind result mismatch (short flag):\n Got: %#v\nWant: %#v", actualShort, expectedShort)
	}
}

// Test binding when optional arguments are not provided.
func TestBindingOptionalArguments(t *testing.T) {
	type OptionalConfig struct {
		RequiredArg  string `docopt:"<req>"`
		OptionalVal  string `docopt:"--opt-val"` // Optional value, defaults to ""
		OptionalFlag bool   `docopt:"-f"`        // Optional flag, defaults to false
		Count        int    `docopt:"--count"`   // Optional int, defaults to 0
	}

	var testParser = &Parser{HelpHandler: NoHelpHandler, SkipHelpFlags: true}
	usage := `Usage: mytool <req> [--opt-val=VAL] [-f] [--count=N]`

	// Case 1: No optional arguments provided
	argvMissing := []string{"required_value"}
	optsMissing, err := testParser.ParseArgs(usage, argvMissing, "")
	if err != nil {
		t.Fatalf("ParseArgs (missing optionals) failed: %v", err)
	}

	expectedMissing := OptionalConfig{
		RequiredArg:  "required_value",
		OptionalVal:  "",    // Expect zero value for string
		OptionalFlag: false, // Expect zero value for bool
		Count:        0,     // Expect zero value for int
	}

	var actualMissing OptionalConfig
	if err := optsMissing.Bind(&actualMissing); err != nil {
		t.Fatalf("Bind (missing optionals) failed: %v", err)
	}

	if !reflect.DeepEqual(actualMissing, expectedMissing) {
		t.Errorf("Bind result mismatch (missing optionals):\n Got: %#v\nWant: %#v", actualMissing, expectedMissing)
	}

	// Case 2: All optional arguments provided
	argvPresent := []string{
		"required_value",
		"--opt-val", "some_value",
		"-f",
		"--count", "42",
	}
	optsPresent, err := testParser.ParseArgs(usage, argvPresent, "")
	if err != nil {
		t.Fatalf("ParseArgs (present optionals) failed: %v", err)
	}

	expectedPresent := OptionalConfig{
		RequiredArg:  "required_value",
		OptionalVal:  "some_value",
		OptionalFlag: true,
		Count:        42,
	}

	var actualPresent OptionalConfig
	if err := optsPresent.Bind(&actualPresent); err != nil {
		t.Fatalf("Bind (present optionals) failed: %v", err)
	}

	if !reflect.DeepEqual(actualPresent, expectedPresent) {
		t.Errorf("Bind result mismatch (present optionals):\n Got: %#v\nWant: %#v", actualPresent, expectedPresent)
	}
}

// Test binding to a struct field that is a pointer.
func TestBindToPointerField(t *testing.T) {
	type PointerConfig struct {
		Name *string `docopt:"--name"`
	}

	var testParser = &Parser{HelpHandler: NoHelpHandler, SkipHelpFlags: true}
	usage := `Usage: mytool --name=NAME`
	argv := []string{"--name", "test"}

	opts, err := testParser.ParseArgs(usage, argv, "")
	if err != nil {
		t.Fatalf("ParseArgs failed: %v", err)
	}

	var actual PointerConfig
	err = opts.Bind(&actual)
	if err == nil {
		t.Fatal("Expected an error when binding to a pointer field, but got nil")
	}

	expectedError := `A pointer field is not supported: "Name".`
	if err.Error() != expectedError {
		t.Errorf("Expected error %q, but got %q", expectedError, err.Error())
	}
}
