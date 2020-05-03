package vm

import (
  "encoding/json"
  "fmt"
  "io/ioutil"
  "math/big"
  "testing"
  "github.com/ethereum/go-ethereum/common"
  "github.com/ethereum/go-ethereum/params"
)

type TwoOperandTestcaseStruct struct {
  X        string
  Y        string
  Expected string
}


var twoOpMethodsMap map[string]executionFunc

func init() {
  twoOpMethodsMap = map[string]executionFunc{
    "add":     opAdd,
    "sub":     opSub,
    "mul":     opMul,
  }
}

func runTwoOperandOp(t *testing.T, tests []TwoOperandTestcaseStruct, opFn executionFunc, name string) {

  var (
    env            = NewEVM(Context{}, nil, params.TestChainConfig, Config{})
    stack          = newstack()
    pc             = uint64(0)
    evmInterpreter = env.interpreter.(*EVMInterpreter)
  )
  // Stuff a couple of nonzero bigints into pool, to ensure that ops do not rely on pooled integers to be zero
  evmInterpreter.intPool = poolOfIntPools.get()
  evmInterpreter.intPool.put(big.NewInt(-1337))
  evmInterpreter.intPool.put(big.NewInt(-1337))
  evmInterpreter.intPool.put(big.NewInt(-1337))
  for i, test := range tests {
    x := new(big.Int).SetBytes(common.Hex2Bytes(test.X))
    y := new(big.Int).SetBytes(common.Hex2Bytes(test.Y))

    expected := new(big.Int).SetBytes(common.Hex2Bytes(test.Expected))
    stack.push(x)
    stack.push(y)
    opFn(&pc, evmInterpreter, &callCtx{nil, stack, nil})
    actual := stack.pop()

    if actual.Cmp(expected) != 0 {
      t.Errorf("Testcase %v %d, %v(%x, %x): expected  %x, got %x", name, i, name, x, y, expected, actual)
    }
    // Check pool usage
    // 1.pool is not allowed to contain anything on the stack
    // 2.pool is not allowed to contain the same pointers twice
    if evmInterpreter.intPool.pool.len() > 0 {

      poolvals := make(map[*big.Int]struct{})
      poolvals[actual] = struct{}{}

      for evmInterpreter.intPool.pool.len() > 0 {
        key := evmInterpreter.intPool.get()
        if _, exist := poolvals[key]; exist {
          t.Errorf("Testcase %v %d, pool contains double-entry", name, i)
        }
        poolvals[key] = struct{}{}
      }
    }
  }
  poolOfIntPools.put(evmInterpreter.intPool)
}

// TestJsonTestcases runs through all the testcases defined as json-files
func TestOpcodes(t *testing.T) {
  for name := range twoOpMethodsMap {
    data, err := ioutil.ReadFile(fmt.Sprintf("testdata/testcasesNew_%v.json", name))
    if err != nil {
      t.Fatal("Failed to read file", err)
    }
    var testcases []TwoOperandTestcaseStruct
    json.Unmarshal(data, &testcases)
    runTwoOperandOp(t, testcases, twoOpMethodsMap[name], name)
  }
}
