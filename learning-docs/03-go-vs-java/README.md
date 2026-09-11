# 3. Go vs Java - A Java Developer's Guide to Go

## Why Go Instead of Java?

| Aspect | Java | Go | Winner for K8s |
|--------|------|-----|----------------|
| **Startup Time** | Slow (JVM warmup) | Fast (native binary) | Go ✓ |
| **Memory Usage** | High (JVM overhead) | Low | Go ✓ |
| **Deployment** | JAR + JVM | Single binary | Go ✓ |
| **Concurrency** | Threads (heavy) | Goroutines (light) | Go ✓ |
| **Kubernetes Libraries** | Third-party | Official | Go ✓ |
| **Learning Curve** | Complex | Simple | Go ✓ |

### The Real Reason: Kubernetes is Written in Go
All official Kubernetes libraries, tools, and examples are in Go. Using Go means:
- Better library support
- More examples to learn from
- Easier integration with K8s ecosystem

---

## Syntax Comparison

### 1. Hello World

**Java:**
```java
public class Main {
    public static void main(String[] args) {
        System.out.println("Hello, World!");
    }
}
```

**Go:**
```go
package main

import "fmt"

func main() {
    fmt.Println("Hello, World!")
}
```

### 2. Variables

**Java:**
```java
int x = 10;
String name = "John";
final int CONSTANT = 100;
```

**Go:**
```go
var x int = 10        // Explicit type
x := 10               // Type inference (short declaration)
name := "John"        // String
const CONSTANT = 100  // Constant
```

### 3. Functions

**Java:**
```java
public int add(int a, int b) {
    return a + b;
}

// Multiple return values? Use a class or array
public int[] divideAndRemainder(int a, int b) {
    return new int[]{a / b, a % b};
}
```

**Go:**
```go
func add(a int, b int) int {
    return a + b
}

// Multiple return values are native!
func divideAndRemainder(a, b int) (int, int) {
    return a / b, a % b
}

// Named return values
func divide(a, b int) (quotient int, remainder int) {
    quotient = a / b
    remainder = a % b
    return // Returns named values
}
```

### 4. Classes vs Structs

**Java (Class with methods):**
```java
public class Person {
    private String name;
    private int age;
    
    public Person(String name, int age) {
        this.name = name;
        this.age = age;
    }
    
    public String getName() {
        return name;
    }
    
    public void greet() {
        System.out.println("Hello, I'm " + name);
    }
}

// Usage
Person p = new Person("John", 30);
p.greet();
```

**Go (Struct with methods):**
```go
type Person struct {
    Name string  // Uppercase = public
    age  int     // Lowercase = private
}

// Constructor function (convention, not language feature)
func NewPerson(name string, age int) *Person {
    return &Person{
        Name: name,
        age:  age,
    }
}

// Method on Person (receiver)
func (p *Person) Greet() {
    fmt.Println("Hello, I'm", p.Name)
}

// Usage
p := NewPerson("John", 30)
p.Greet()
```

### 5. Interfaces

**Java:**
```java
public interface Animal {
    void speak();
}

public class Dog implements Animal {
    @Override
    public void speak() {
        System.out.println("Woof!");
    }
}
```

**Go (Implicit interfaces - no "implements" keyword!):**
```go
type Animal interface {
    Speak()
}

type Dog struct{}

// Dog automatically implements Animal because it has Speak()
func (d Dog) Speak() {
    fmt.Println("Woof!")
}

// Usage
var animal Animal = Dog{}
animal.Speak()
```

### 6. Error Handling

**Java (Exceptions):**
```java
public String readFile(String path) throws IOException {
    // May throw IOException
    return Files.readString(Path.of(path));
}

// Usage
try {
    String content = readFile("file.txt");
} catch (IOException e) {
    System.err.println("Error: " + e.getMessage());
}
```

**Go (Explicit error returns):**
```go
func readFile(path string) (string, error) {
    content, err := os.ReadFile(path)
    if err != nil {
        return "", err  // Return error
    }
    return string(content), nil  // nil = no error
}

// Usage
content, err := readFile("file.txt")
if err != nil {
    fmt.Println("Error:", err)
    return
}
fmt.Println(content)
```

### 7. Inheritance vs Composition

**Java (Inheritance):**
```java
public class Animal {
    protected String name;
    public void eat() { System.out.println("Eating..."); }
}

public class Dog extends Animal {
    public void bark() { System.out.println("Woof!"); }
}
```

**Go (Composition - No inheritance!):**
```go
type Animal struct {
    Name string
}

func (a *Animal) Eat() {
    fmt.Println("Eating...")
}

// Dog "embeds" Animal (composition)
type Dog struct {
    Animal  // Embedded struct
}

func (d *Dog) Bark() {
    fmt.Println("Woof!")
}

// Usage
dog := Dog{Animal: Animal{Name: "Buddy"}}
dog.Eat()   // Works! Promoted from Animal
dog.Bark()  // Dog's own method
```

### 8. Null vs Nil

**Java:**
```java
String s = null;
if (s == null) {
    System.out.println("s is null");
}
```

**Go:**
```go
var s *string = nil  // Pointer to string, nil
if s == nil {
    fmt.Println("s is nil")
}

// Slices, maps, channels can also be nil
var slice []int = nil
var m map[string]int = nil
```

### 9. Collections

**Java:**
```java
// List
List<String> list = new ArrayList<>();
list.add("a");
list.add("b");

// Map
Map<String, Integer> map = new HashMap<>();
map.put("one", 1);
map.put("two", 2);
```

**Go:**
```go
// Slice (like ArrayList)
slice := []string{"a", "b"}
slice = append(slice, "c")

// Map
m := map[string]int{
    "one": 1,
    "two": 2,
}
m["three"] = 3

// Check if key exists
value, exists := m["one"]
if exists {
    fmt.Println(value)
}
```

### 10. Concurrency

**Java:**
```java
// Thread
Thread t = new Thread(() -> {
    System.out.println("Running in thread");
});
t.start();
t.join();
```

**Go:**
```go
// Goroutine (much lighter than threads)
go func() {
    fmt.Println("Running in goroutine")
}()

// Channels for communication
ch := make(chan string)
go func() {
    ch <- "Hello from goroutine"
}()
msg := <-ch  // Receive from channel
fmt.Println(msg)
```

---

## Key Differences Summary

| Feature | Java | Go |
|---------|------|-----|
| OOP | Classes, inheritance | Structs, composition |
| Interfaces | Explicit (`implements`) | Implicit (duck typing) |
| Errors | Exceptions | Return values |
| Null | `null` | `nil` |
| Generics | Yes (since Java 5) | Yes (since Go 1.18) |
| Visibility | `public/private/protected` | Uppercase/lowercase |
| Constructors | Constructor methods | Factory functions |
| Getters/Setters | Common pattern | Not idiomatic |
| Package management | Maven/Gradle | Go modules |

---

## Go Idioms to Learn

### 1. The Blank Identifier `_`
```go
// Ignore a return value
_, err := someFunction()

// Import for side effects only
import _ "some/package"
```

### 2. Defer
```go
func readFile() {
    file, _ := os.Open("file.txt")
    defer file.Close()  // Will run when function returns
    
    // ... use file
}
```

### 3. Type Assertions
```go
var i interface{} = "hello"

s := i.(string)        // Panic if not string
s, ok := i.(string)    // Safe - ok is false if not string
```

### 4. Type Switch
```go
switch v := i.(type) {
case string:
    fmt.Println("string:", v)
case int:
    fmt.Println("int:", v)
default:
    fmt.Println("unknown type")
}
```

---

## Next Steps

Continue to [Kubernetes Fundamentals](../04-kubernetes-fundamentals/README.md) to understand the Kubernetes concepts used in this operator.
