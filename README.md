# Zap-Lipgloss
Zap-Lipgloss is a custom encoder for [zap](https://github.com/uber-go/zap) that uses [Charmbracelet/Lipgloss](https://github.com/charmbracelet/lipgloss/) to create a beautiful readable format.

![image](https://github.com/user-attachments/assets/56b1c68d-1fb8-40b0-8bc4-136c228aea3f)

```go
package main

import (
    _ "github.com/tvanriel/zaplipgloss" // add the package as _ to register the logger.
	"go.uber.org/zap"
)

func main() {
  logger, err := zap.Config{
    Level: zap.NewAtomicLevelAt(zap.DebugLevel),
    Encoding: "lipgloss", // Set the logger in the config
    OutputPaths: []string{"stdout"},
  }.Build()

}
```
