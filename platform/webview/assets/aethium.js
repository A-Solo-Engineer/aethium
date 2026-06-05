// Minimal Aethium host bridge for testing
window.Aethium = {
  initRuntime: function(canvasID) {
    console.log('Aethium runtime initialized for canvas', canvasID);
  },
  pumpEvents: function() {},
  renderFrame: function() {}
};
