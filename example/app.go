package main

import (
	_ "github.com/tvanriel/zaplipgloss"
	"go.uber.org/zap"
)

func main() {
  logger, err := zap.Config{
    Level: zap.NewAtomicLevelAt(zap.DebugLevel),
    Encoding: "lipgloss",
    OutputPaths: []string{"stdout"},
  }.Build()
  if err != nil {
    panic(err)
  }
  logger.Info("test",
    zap.Int("one", 1),
    zap.String("Rainbows!", "omg"),
    zap.String("Rainbows!", "omg"),
    zap.String("Rainbows!", "omg"),
    zap.String("Rainbows!", "omg"),
    zap.String("Rainbows!", "omg"),
    zap.String("Rainbows!", "omg"),
    zap.String("Rainbows!", "omg"),
    zap.String("Rainbows!", "omg"),
  )
  
  logger.Warn("warn")
  logger.Debug("debug")
  logger.DPanic("d_panic")
  logger.Fatal("fatal")
  logger.Error("error")
  logger.Panic("panic")
}
