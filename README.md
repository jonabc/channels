# Channels

Some utilities for working with channels.

While this library is useful for creating chained operations into a pipeline, it is intended to be lightweight. Pipeline orchestration, error handling, logging, and metrics tracking should be implemented by callers.

## Types

### Provider[T any] and Receiver[T any]

Providers and receivers wrap channels to provide a few quality of life benefits.
1. Writing to the provider while it is closing or after it is closed will not produce a panic
2. A provider can be configured for different behaviors on calling `Provide`
   - `providers.NewProvider` matches the underlying channel behavior, blocking when the receiving channel blocks
   - `providers.NewDroppingProvider` drops provided values when the receiving channel blocks
   - `providers.NewCollectingProvider` collects observed values while the underlying channel blocks.  When the receiving channel is unblocked all values are written to the receiver as a slice.

## Functions

### Batch

```go
// signature
func Batch[T any](inc <-chan T, batchSize int) <- chan T

// usage
inc := make(chan int)
outc := Batch(inc, 2)

inc <- 1
inc <- 2
close(inc)

results := <- outc
// results == []int{1,2}
```

Batch N values from the input channel into an array of N values in the output channel.  The output channel is unbuffered by default, and will be closed when the input channel is closed and drained.  If a partial batch exists when the input channel is closed, the partial batch will be sent to the output channel.

### BatchValues (Blocking)

```go
// signature
func BatchValues[T any](inc <-chan T, batchSize int) [][]T

// usage
inc := make(chan int, 4)
inc <- 1
inc <- 2
inc <- 3
inc <- 4
close(inc)

results := BatchValues(inc, 2)
// results == [][]int{{1,2}, {3,4}}
```

Like Batch, but blocks until the input channel is closed and all values are read.  BatchValue reads all values from the input channel and returns an array of batches.

### Debounce

```go
// signature
func Debounce[T comparable](inc <-chan T, delay time.Duration) (<- chan T, func() int)

// usage
inc := make(chan int)
defer close(inc)

delay := 5 * time.Second
outc := Debounce(inc, delay)

// writing to the input channel starts a debouncing period
inc <- 1

// a debouncing period is active, both of these values are ignored
inc <- 1
inc <- 2

// will block for approx `delay` time period
results := <- outc

// results == 1
// any future writes to inc will start a new debouncing period
```

Debounce reads values from the input channel and pushes them to the returned output channel before, after, or before and after a `delay` debounce period. The [DebounceType](#debounce-types) value used in the function controls when the value is pushed to the output channel.  Any duplicate values read from the input channel during the `delay` debounce period are ignored.

The channel returned by Debounce is unbuffered by default.  When the input channel is closed, any remaining values being delayed/debounced will be flushed to the output channel and the output channel will be closed.

Debounce also returns a function which returns the number of debounced values that are currently being delayed

For more complicated use cases, see [DebounceCustom](./#debouncecustom) below.

#### Debounce Types

When debouncing values, the value can either be written to the output channel before the debounce period starts (`LeadDebounceType`), after the debounce period ends (`TailDebounceType`) or both (`LeadTailDebounceType`).  The default is `TailDebounceType`.

The debounce type can be set using the `DebounceTypeOption` function option.


### DebounceValues

```go
// signature
func DebounceValues[T comparable](inc <-chan T, delay time.Duration) (<- chan T, func() int)

// usage
inc := make(chan int)
defer close(inc)

delay := 5 * time.Second
outc := Debounce(inc, delay)

// this write starts a debounce period for the value 1
inc <- 1
time.Sleep(1 * time.Second)

// this write duplicates the previous value 1 written to inc
// and is ignored
inc <- 1

// this write starts a debounce period for the value 2
inc <- 2

// will block for approx `delay - 1s` time period
results := <- outc

// will block for approx 1 second
results += <- outc
// results == 3, len(outc) == 0
```

DebounceValues is like [Debounce](#debounce) but with per-value debouncing, where each unique value read from the input channel will start a debouncing period for that value.  Any duplicate values read from the input channel during a debouncing period are ignored.

### DebounceCustom

```go
// signature
type Keyable[K comparable] interface {
	Key() K
}

type DebounceInput[K comparable, T Keyable[K]] interface {
	Keyable[K]
	Delayable
	Reduce(T) (T, bool)
}

func DebounceCustom[K comparable, T DebounceInput[K, T]](inc <-chan T) (<- chan T, func() int)

// usage
type myType struct {
	key   string
	value int
	delay time.Duration
}

func (d *myType) Key() string {
	return d.key
}

func (d *myType) Delay() time.Duration {
	return d.delay
}

func (d *myType) Reduce(other *myType) (*myType, bool) {
  if d.key == "3" {
    return d, false
  }

  d.value += other.value
	return d, true
}

inc := make(chan *myType)
defer close(inc)

delay := 5 * time.Second
outc := DebounceCustom(inc)

inc <- &myType{key:"1", val: 1, delay: delay}

time.Sleep(1 * time.Second)
inc <- &myType{key:"2", val: 2, delay: delay}
inc <- &myType{key:"1", val: 3, delay: 1 * time.Millisecond}
inc <- &myType{key:"3", val: 3, delay: delay}
inc <- &myType{key:"3", val: 4, delay: 1 * time.Millisecond}

// will block for approx `delay - 1s` time period
result := <- outc
// result == &myType{key:"1", val: 4, delay: delay}

result = <- outc
// result == &myType{key:"2", val: 2, delay: delay}

result = <- outc
// result == &myType{key: "3", val: 3, delay: delay}
```

DebounceCustom is like [Debounce](#debounce) but with per-item configurability over comparisons, delays, and reducing multiple values to a single debounced value.  Where Debounce requires `comparable` values in the input channel, DebounceCustom requires types that implement the `DebounceInput[K comparable, T Keyable[K]]` interface.  The typing is a little complex but in practice most of the complexity is hidden from callers.  Values of type `T` passsed into DebounceCustom must implement:
1. `Key() K` returns a `comparable` value
   - The key is used to compare values read from the input channel for uniqueness.  If multiple values with the same key are seen, they will be reduced into a single value.
2. `Delay() time.Duration` returns the debounce delay period for the value
   - This function is only called for the first value debounced for each unique key, i.e. if a value is read from the input channel with the same `Key()` as an existing delayed value, the delay from the previously seen value is maintained
3. `Reduce(T) T` combines the value with another value.  As duplicate values are seen (as determined by comparisons of `Key()`), they will be continuously reduced to a single value which will be returned after the debounce period for that value has elapsed.

### Delay

```go
// signature
func Delay[T any](inc <-chan T, delay time.Duration) (<- chan T, func() int)

// usage
inc := make(chan int)
defer close(inc)

delay := 5 * time.Second
outc := Delay(inc, delay)

// writing to the input channel starts a delay period
inc <- 1

// will block for approx `delay` time period
results := <- outc

// results == 1
```

Delay reads values from the input channel, and writes each value to the output channel after `delay` duration.

The channel returned by Delay is unbuffered by default.  When the input channel is closed, any remaining values being delayed will be flushed to the output channel and the output channel will be closed.

Delay also returns a function which returns the number of values that are currently being delayed.

For more complicated use cases, see [DelayCustom](./#delaycustom) below.

### DelayCustom

```go
// signature
type Delayable interface {
	Delay() time.Duration
}

func DelayCustom[T Delayable](inc <-chan T) (<- chan T, func() int)

// usage
type myType struct {
	delay time.Duration
}

func (d *myType) Delay() time.Duration {
	return d.delay
}

inc := make(chan *myType)
defer close(inc)

delay := 5 * time.Second
outc := DelayCustom(inc)

inc <- &myType{delay: delay}

// will block for approx `delay` time period
result := <- outc
// result == &myType{delay: delay}
```

DelayCustom is like [Delay](#delay) but with per-item configurability over delays.  DelayCustom requires types that implement the `Delayable` interface.

### Drain (Blocking)

```go
// signature
func Drain[T any](inc <-chan TIn, maxWait time.Duration) (int, bool)

// usage
inc := make(chan int)
inc <- 1
inc <- 2
inc <- 3
close(inc)

count, drained := channels.Drain(inc, 10 * time.Millisecond)
// count == 3, drained == true
```

Drain blocks until either the input channel is fully drained and closed or `maxWait` duration has passed.  Drain returns the count of values drained from the channel, and a bool that is true when exiting due to the input channel being drained and closed or false when exiting due to waiting for the `maxWait` duration.  When `maxWait <= 0`, Drain will wait forever, and only exit when the input channel is closed.

### DrainValues (Blocking)

```go
// signature
func DrainValues[T any](inc <-chan TIn, maxWait time.Duration) ([]TIn, bool)

// usage
inc := make(chan int)
inc <- 1
inc <- 2
inc <- 3
close(inc)

values, drained := channels.Drain(inc, 10 * time.Millisecond)
// valued == []int{1,2,3}, drained == true
```

DrainValues blocks until either the input channel is fully drained and closed or `maxWait` duration has passed.  DrainValues returns the values drained from the channel, and a bool that is true when exiting due to the input channel being drained and closed or false when exiting due to waiting for the `maxWait` duration.  When `maxWait <= 0`, Drain will wait forever, and only exit when the input channel is closed.

### Each

```go
// signature
func Each[TIn any](inc <-chan TIn, eachFn func(TIn))

// usage
inc := make(chan int)
// map integers to a boolean indicating if the values are odd (false) or even (true)
outc := Each(inc, func(i int) (bool, bool) { return i%2 == 0, i < 3 })

inc <- 1
inc <- 2
inc <- 3
close(inc)

results := []int{}
for result := range outc {
  result = append(results, result)
}
// results == []bool{false, true}
```

Each consumes values from the input channel and applies the provided `eachFn` to each value.  Values are not propagated to an output channel, consider using [Tap](#tap) or [Map](#map) for channel propagation.

### FlatMap

```go
// signature
func FlatMap[TIn any, TOut any, TOutSlice []T](inc <-chan TIn, mapFn func(TIn) (TOut, bool)) <- chan TOut

// usage
inc := make(chan int)
// map integers to an array of different values
outc := FlatMap(inc, func(i int) ([]int, bool) { return []int{i*10,i*10+1}, i < 3 })

inc <- 1
inc <- 2
inc <- 3
close(inc)

results := []int{}
for result := range outc {
  result = append(results, result)
}
// results == []int{10,11,20,21}
```

FlatMap reads values from the input channel and applies the provided `mapFn` to each value.  Each element in the slice returned by `mapFn` is then sent to the output channel.

The output channel is unbuffered by defualt, and is closed once the input channel is closed and all mapped values are pushed to the output channel.

### FlatMapValues (Blocking)

```go
// signature
func FlatMapValues[TIn any, TOut any](inc <-chan TIn, mapFn func(TIn) (TOut, bool)) []TOut

// usage
inc := make(chan int)

inc <- 1
inc <- 2
inc <- 3
close(inc)

// map integers to an array of different values
results := FlatMapValues(inc, func(i int) ([]int, bool) { return []int{i*10,i*10+1}, i < 3 })
// results == []int{10, 11, 20, 21}
```

Like FlatMap, but blocks until the input channel is closed and all values are read.  FlatMapValues reads all values from the input channel and returns a flattened array of values returned from passing each input value into `mapFn`.

### Map

```go
// signature
func Map[TIn any, TOut any](inc <-chan TIn, mapFn func(TIn) (TOut, bool)) <- chan TOut

// usage
inc := make(chan int)
// map integers to a boolean indicating if the values are odd (false) or even (true)
outc := Map(inc, func(i int) (bool, bool) { return i%2 == 0, i < 3 })

inc <- 1
inc <- 2
inc <- 3
close(inc)

results := []int{}
for result := range outc {
  result = append(results, result)
}
// results == []bool{false, true}
```

Map reads values from the input channel and applies the provided `mapFn` to each value before pushing it to the output channel.  The output channel is unbuffered by default, and will be closed once the input channel is closed and all mapped values pushed to the output channel.  The type of the output channel does not need to match the type of the input channel.

### MapValues (Blocking)

```go
// signature
func MapValues[TIn any, TOut any](inc <-chan TIn, mapFn func(TIn) (TOut, bool)) []TOut

// usage
inc := make(chan int)

inc <- 1
inc <- 2
inc <- 3
close(inc)

// map integers to a boolean indicating if the values are odd (false) or even (true)
results := Map(inc, func(i int) (bool, bool) { return i%2 == 0, i < 3 })
// results == []bool{false, true}
```

Like Map, but blocks until the input channel is closed and all values are read.  MapsValues reads all values from the input channel and returns an array of values returned from passing each input value into `mapFn`.

### Merge

```go
// signature
func Merge[T any](chans ...<-chan T) <-chan T

// usage

inc1 := make(chan int)
inc2 := make(chan int)
inc3 := make(chan int)

outc := Merge([]chan int{inc1, inc2, inc3})

var wg sync.WaitGroup
wg.Add(1)
results := []int{}
go func() {
  defer wg.Done()
  for val := range outc {
    results = append(results, val)
  }
}()

inc1 <- 1
inc2 <- 2
inc3 <- 3
inc1 <- 4

close(inc1)
close(inc2)
close(inc3)

wg.Wait()
// results will have the same elements as []int{1,2,3,4} but may not be in that order
```

Merge merges multiple input channels into a single output channel.  The order of values in the output channel is not guaranteed to match the order that values are written to the input channels.  The output channel is unbuffered by default and is closed when all input channels are closed.

### Reduce

```go
// signature
func Reduce[TIn any, TOut any](inc <-chan T, reduceFn func(TOut, TIn) (TOut, bool)) <- chan TOut

// usage
inc := make(chan int)
// reduce integer values into an array of string values,
// creating a new array on every iteration.  see below for
// the difference if `return append(...)` is used instead
outc := Reduce(inc, func(accum []string, i int) string { 
  output := make([]string, len(accum) + 1)
  copy(output, current)
  output[len(output)-1] = strconv.Itoa(i)
  
  return output
})

inc <- 1
inc <- 2
close(inc)

results := []int{}
for result := range outc {
  result = append(results, result)
}
// results == [][]string{{"1"}, {"1", "2"}}

// NOTE:
// If the reducer returned `append(accum, strconv.Itoa(i))`,
// the same array would be pushed to the output channel multiple times
// and the result will be [][]string{{"1", "2"}, {"1", "2"}}
```

Reduce reads values from the input channel and applies the provided `reduceFn` to each value.  The first argument to the reducer function is the accumulated reduced value, which is either (a) the default value for the type on the first call or (b) the output from the previous call of the reducer function for all other iterations.

The output of each call to the reducer function is pushed to the output channel.

### ReduceValues (Blocking)

```go
// signature
func ReduceValues[TIn any, TOut any](inc <-chan T, reduceFn func(TOut, TIn) (TOut, bool)) TOut

// usage
inc := make(chan int, 2)
inc <- 1
inc <- 2
close(inc)

// reduce integers to an array of strings
results := ReduceValues(inc, func(accum []string, i int) bool { return append(accum, strconv.Itoa(i)) })
// results == []{"1", "2"}
```

Like Reduce, but blocks until the input channel is closed and all values are read.  ReduceValues reads all values from the input channel and returns the value returned after all values from the input channel have been passed into `reduceFn`.

### Reject

```go
// signature
func Reject[T any](inc <-chan T, rejectFn func(T) bool) <- chan T

// usage
inc := make(chan int)
// reject even numbers
outc := Reject(inc, func(i int) bool { return i%2 == 0 })

inc <- 1
inc <- 2
close(inc)

results := []int{}
for result := range outc {
  result = append(results, result)
}
// results == []int{1}
```

Selects values from the input channel that return false from the provided `rejectFn` and pushes them to the output channel.  The output channel is unbuffered by default, and is closed once the input channel is closed and all selected values pushed to the output channel.

### RejectValues (Blocking)

```go
// signature
func RejectValues[T any](inc <-chan T, rejectFn func(T) bool) []T

// usage
inc := make(chan int, 4)
inc <- 1
inc <- 2
inc <- 3
inc <- 4
close(inc)

result := RejectValues(inc, func(i int) bool { return i%2 == 0 })
// results == []int{1, 3}
```

Like Reject, but blocks until the input channel is closed and all values are read.  RejectValues reads all values from the input channel and returns an array of values that return false from the provided `rejectFn` function.

### Select

```go
// signature
func Select[T any](inc <-chan T, selectFn func(T) bool) <- chan T

// usage
inc := make(chan int)
// select even numbers only
outc := Select(inc, func(i int) bool { return i%2 == 0 })

inc <- 1
inc <- 2
close(inc)

results := []int{}
for result := range outc {
  result = append(results, result)
}
// results == []int{2}
```

Selects values from the input channel that return true from the provided `selectFn` and pushes them to the output channel.  The output channel is unbuffered by default, and is closed once the input channel is closed and all selected values pushed to the output channel.


### SelectValues (Blocking)

```go
// signature
func SelectValues[T any](inc <-chan T, selectFn func(T) bool) []T

// usage
inc := make(chan int, 4)
inc <- 1
inc <- 2
inc <- 3
inc <- 4
close(inc)

result := SelectValues(inc, func(i int) bool { return i%2 == 0 })
// results == []int{2, 4}
```

Like Select, but blocks until the input channel is closed and all values are read.  SelectValues reads all values from the input channel and returns an array values that return true from the provided `selectFn` function.

### Split

```go
// signature
func Split[T any](inc <-chan T, count int, splitFn func(T, []chan<- T)) []<-chan T

// usage
inc := make(chan int, 4)
// split incoming values into separate chanenls for even and odd values
outChans := channels.Splt(inc, 2, func(i int, chans []chan<- int) {
  chans[i%2] <- i
})

inc <- 1
inc <- 2
inc <- 3
inc <- 4

odds := []int{}
for result := range outChans[1] {
  odds = append(odds, result)
}
// odds == []int{1,3}

evens := []int{}
for result := range outChans[0] {
  evens = append(evens, result)
}
// evens == []int{2,4}
```

Split reads values from the input channel and routes the values into `N` output channels using the provided `splitFn`.  The channel slice provided to `splitFn` will have the same length and order as the channel slice returned from the function, e.g. in the above example `Split` guarantees that chans[0] will hold even values and chans[1] will hold odd values.

Each output channel is unbuffered by default, and will be closed after the input channel is closed and emptied.

### SplitValues (Blocking)

```go
// signature
func SplitValues[T any](inc <-chan T, count int, splitFn func(T, []chan<- T)) [][]T

// usage
inc := make(chan int, 4)
inc <- 1
inc <- 2
inc <- 3
inc <- 4
close(inc)

results := SplitValues(inc, func(i int, chans []chan<- T) { chans[i%2] <- i })
evens := results[0]
odds := results[1]
// results == [][]int{{2, 4}, {1, 3}}
// evens == []int{2, 4}
// odds == []int{1, 3}
```

Like Split, but blocks until the input channel is closed and all values are read.  SplitValues reads all values from the input channel and returns `[][]T`, a two-dimensional slice containing the results from each split channel.
- The first dimension, `i` in `[i][j]T` matches the size and order of channels provided to `splitFn`
- The second dimension, `j` in `[i][j]T` matches the size and order of values written to `chans[i]` in `splitFn`

### Tap

```go
// signature
func Tap[T any](inc <-chan T, preFn func(T), postFn func(T)) <-chan T

// usage

inc := make(chan int, 4)
// log values to stdout after they are written to the output channel
outc := channels.Tap(inc, nil, func(i int) { fmt.Println(i) })

inc <- 1
inc <- 2
close(inc)

results := []int{}
for result := range outc {
  result = append(results, result)
}
// results == []int{1, 2}
```

Tap reads values from the input channel and calls the provided `[pre/post]Fn` functions with each value before and after writing the value to the output channel, respectivel.  The output channel is unbuffered by default, and will be closed after the input channel is closed and drained.

### ThrottleValues

```go
// signature
func Throttle[T comparable](inc <-chan T, delay time.Duration) (<- chan T, func() int)

// usage
inc := make(chan int)
defer close(inc)

delay := 5 * time.Second
outc := Throttle(inc, delay)

// writing to inc will result in a value published to outc immediately
inc <- 1
<-outc

// all writes to inc during the throttling period will not result in any writes to outc
inc <- 1
inc <- 2

// wait for throttle period to end
time.Sleep(delay + 1*time.Millisecond)

// writing to inc will result in a value published to outc immediately
inc <- 1
<-outc
```

Throttle is equivalent to [Debounce](#debounce) with `channels.LeadDebounceType`.

### ThrottleValues

```go
// signature
func ThrottleValues[T comparable](inc <-chan T, delay time.Duration) (<- chan T, func() int)

// usage
inc := make(chan int)
defer close(inc)

delay := 5 * time.Second
outc := ThrottleValues(inc, delay)

// writing to inc will result in a value published to outc immediately
inc <- 1
<-outc

// duplicate write of 1 to inc is throttled and will not result in any writes to outc
inc <- 1

// writing a new value to inc will result in the value published to outc immediately
inc <- 2
<-outc

// wait for throttle period to end
time.Sleep(delay + 1*time.Millisecond)

// writing to inc will result in a value published to outc immediately
inc <- 1
<-outc
```

ThrottleValues is equivalent to [DebounceValues](#debouncevalues) with `channels.LeadDebounceType`.

### ThrottleCustom

```go
// signature
type Keyable[K comparable] interface {
	Key() K
}

type DebounceInput[K comparable, T Keyable[K]] interface {
	Keyable[K]
	Delay() time.Duration
	Reduce(T) (T, bool)
}

func ThrottleCustom[K comparable, T DebounceInput[K, T]](inc <-chan T) (<- chan T, func() int)

// usage
type myType struct {
	key   string
	value int
	delay time.Duration
}

func (d *myType) Key() string {
	return d.key
}

func (d *myType) Delay() time.Duration {
	return d.delay
}

func (d *myType) Reduce(other *myType) (*myType, bool) {
  return d, true
}

inc := make(chan *myType)
defer close(inc)

delay := 5 * time.Second
outc := ThrottleCustom(inc)

// writing to inc will result in a value published to outc immediately
inc <- &myType{key:"1", val: 1, delay: delay}
<-outc

// write to inc is throttled and will not result in any writes to outc
inc <- &myType{key:"1", val: 1, delay: delay}

// wait for throttle period to end
time.Sleep(delay + 1*time.Millisecond)

// writing to inc will result in a value published to outc immediately
inc <- &myType{key:"1", val: 1, delay: delay}
<-outc
```

ThrottleCustom is equivalent to [DebounceCustom](#debouncecustom) with `channels.LeadDebounceType`.

### Unique

```go
// signature
func Unique[T comparable](inc <-chan T, batchSize int, maxDelay time.Duration) <- chan T

// usage
inc := make(chan int)
defer close(inc)

outc := Unique(inc, 2, 0)

inc <- 1
inc <- 1
inc <- 2

results := <- outc
// results == []int{1,2}
```

Batch unique values from the input channel into an array of values written to the output channel.

The output channel is unbuffered by default, and will be closed when the input channel is closed and drained.  If a partial batch exists when the input channel is closed, the partial batch will be sent to the output channel.

### UniqueKeyed

```go
// signature
type Keyable[K comparable] interface {
	Key() K
}

func UniqueKeyed[K comparable, V Keyable[K]](inc <-chan V, batchSize int, maxDelay time.Duration) <-chan []V

// usage
type myType struct {
	key   string
  value int
}

func (d *myType) Key() string {
	return d.key
}

inc := make(chan *myType)
defer close(inc)

outc := UniqueKeyed(inc, 2, 0)

inc <- &myType{key:"1", val: 1}
inc <- &myType{key:"1", val: 1}
inc <- &myType{key:"2", val: 2}

result := <- outc
// result == []*myType{&myType{key:"1", val: 1}, &myType{key:"2", val: 2}}
```

Batch unique values from the input channel into an array of values written to the output channel.  Uniqueness is determed by the value returned by each value's Key() function.

The output channel is unbuffered by default, and will be closed when the input channel is closed and drained.  If a partial batch exists when the input channel is closed, the partial batch will be sent to the output channel.

### UniqueValues

```go
// signature
func UniqueValues[T any](inc <-chan T, batchSize int, maxDelay time.Duration) [][]T

// usage
inc := make(chan int, 4)
inc <- 1
inc <- 1
inc <- 2
inc <- 3
inc <- 3
inc <- 4
close(inc)

results := UniqueValues(inc, 2, 0)
// results == [][]int{{1,2}, {3,4}}
```

Like Unique, but blocks until the input channel is closed and all values are read.  UniqueValues reads all values from the input channel and returns an array of batches.

### Void

```go
// signature
Void[T any](inc <-chan T)

// usage
inc := make(chan int, 2)
channels.Void(inc)

inc <- 1
inc <- 2

// and they're gone
```

Void consumes a channel with no other action taken.  This function can be useful for tests or while iteratively composing a channel pipeline before the final consumption of channel events is completed.

### WithDone

```go
// signature
WithDone[T any](inc <-chan T) (<-chan T, <-chan struct{})

// usage
inc := make(int, 2)
outc, done := channels.WithDone(inc)

go func() {
  for {
    select {
    case <-done:
      fmt.Println("Finished")
      return
    default:
      fmt.Printf("Channel currently has %d items\n", len(outc))
      time.Sleep(10 * time.Millisecond)
    }
  }
}()

// do things...

close(inc)
```

WithDone returns two channels: a channel containing piped input from the input channel as well and a channel which will be closed when the input channel has been closed and all values written to the piped output channel.

WithDone is meant to be used in situations where a component needs awareness of the lifetime of a channel but interacting with the channel directly is not desirable.  In the example above, the `done` channel is used in a goroutine to report the current length of the channel at a regular interval.

## Options

###  Specifying output channel capacity (single channel output)

Functions returning a single output channel set a capacity calculated from the input channel; see each function description for default output channel capacities.  The output channel capacity can be overridden by passing `channels.ChannelCapacityOption` to the function.

```go
// signature
channels.ChannelCapacityOption[T singleOutputConfiguration](capacity int)

// usage
inc := make(chan int, 10)
defer close(inc)

// map integers to a boolean indicating if the values are odd (false) or even (true)
outc := Map(inc, 
  func(i int) (bool, bool) { return i%2 == 0, i < 3 },
  channels.ChannelCapacityOption[channels.MapConfig](1),
)

// cap(outc) == 1
```

### Specifying output channels' capacities (multi channel output)

Functions returning multiple output channels set capacities onto the channels calculated from the input channel; see each function description for default output channel capacities.  The output channel capacities can be overridden by passing `channels.ChannelCapacityOption` to the function.

```go
// signature
channels.MultiChannelCapacitiesOption[T multiOutputConfiguration](capacities []int)

// usage
inc := make(chan int, 10)
defer close(inc)

outcs := Split(inc, 2,
  func(i int, chans []chan<- int) { chans[i%2] <- i },
  channels.MultiChannelCapacitiesOption[channels.SplitConfig]([]int{3,4}),
)

// cap(outcs[0]) == 3
// cap(outcs[0]) == 4
```

### Specifying a provider for panic reporting

Most of the behavior in this package happens in goroutines, and a panic can cause an application to crash without triggering any configured logging or graceful error handling behaviors.  Panics can be captured and proxied to callers by creating an `providers.Provider[any]` and passing it to the function via `channels.PanicProviderOption`.

```go
// signature
channels.PanicProviderOption[T channelConfiguration](providers.Provider[any]) Option[T]

// usage
inc := make(chan int, 10)
defer close(inc)

panicProvider, panicReceiver := providers.NewProvider[any](0)
defer panicProvider.Close()

// map integers to a boolean indicating if the values are odd (false) or even (true)
outc := Map(inc, 
  func(i int) (bool, bool) { panic("oops") },
  channels.PanicProviderOption[channels.MapConfig](panicProvider),
)

inc <- 1
// "oops" == <- panicReceiver.Channel()
```

### Specifying a provider for stats reporting

Most channel functions take an options argument that allows callers to receive information about the channel function's operations over time.  While it is possible to manually observe most channel operations using `channels.Tap` to observe items moving through a channel pipeline, using providers to report on stats provides a couple of additional benefits:
1. If needed, providers can be closed to stop stats reporting without otherwise affecting a channel's operations.
2. Using the collecting provider provides a fully asynchronous way to gather statistics about a channel operation, in case reporting stats would otherwise affect channel throughput.

```go
inc := make(chan int, 10)
defer close(inc)

statsProvider, statsReceiver := providers.NewCollectingProvider[Stats](0)
defer statsProvider.Close()

go func() {
  for stats := range statsReceiver.Channel() {
    for i, stat := range stats {
      // report on stat.Duration to an observability tool
    }
  }
}()

out := channels.Map(in, 
  func(i int) (bool, bool) {return i%2==0, true},
  channels.StatsProviderOption[channels.MapConfig](statsProvider),
)

// ...
```

#### Choosing a provider for reporting statistics

If you need high fidelity in statistics reporting, `providers.NewCollectingProvider` will gather statistics in memory when the receiving channel is blocked.  When the receiving channel is unblocked, the next read will contain all statistics collected in memory while the channel was blocked.

If you do not need high fidelity reporting and want to optimize for CPU and memory usage, `providers.NewDroppingProvider` will drop any provided statistics whenever the receiving channel is blocked.

**Warning**  It's generally recommended to not using the blocking provider `providers.NewProvider` to receive channel statistics, as this provider will block a channel operation until the receiving channel can be written to.