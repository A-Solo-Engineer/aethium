// Aethium host bridge for WebAssembly integration
window.Aethium = {
  // Runtime management
  initRuntime: function(canvasID) {
    console.log('Aethium runtime initialized for canvas', canvasID);
    this.canvasID = canvasID;
    this.canvas = document.getElementById(canvasID);
    if (!this.canvas) {
      console.error('Canvas with ID', canvasID, 'not found');
      return;
    }
    
    // Set up canvas context
    this.ctx = this.canvas.getContext('2d');
    if (!this.ctx) {
      console.error('Failed to get 2D context for canvas', canvasID);
      return;
    }
    
    // Initialize animation frame
    this.animationId = null;
    this.isRunning = false;
    
    console.log('Aethium runtime setup complete for canvas', canvasID);
  },
  
  // Event handling
  pumpEvents: function() {
    // Handle browser events and convert to Aethium events
    // This would be implemented based on the specific event system
    console.log('Pumping events...');
  },
  
  // Rendering
  renderFrame: function(drawCommands) {
    if (!this.ctx) {
      console.error('Canvas context not available');
      return;
    }
    
    // Clear canvas
    this.ctx.clearRect(0, 0, this.canvas.width, this.canvas.height);
    
    // Process draw commands
    for (var i = 0; i < drawCommands.length; i++) {
      var cmd = drawCommands[i];
      this.processDrawCommand(cmd);
    }
  },
  
  // Draw command processing
  processDrawCommand: function(cmd) {
    switch (cmd.kind) {
      case 1: // CmdFillRect
        this.ctx.fillStyle = this.argbToHex(cmd.color);
        this.ctx.fillRect(cmd.x, cmd.y, cmd.w, cmd.h);
        break;
      case 2: // CmdStrokeRect
        this.ctx.strokeStyle = this.argbToHex(cmd.color);
        this.ctx.strokeRect(cmd.x, cmd.y, cmd.w, cmd.h);
        break;
      case 3: // CmdDrawText
        this.ctx.fillStyle = this.argbToHex(cmd.color);
        this.ctx.font = '16px Arial';
        this.ctx.fillText(cmd.text, cmd.x, cmd.y);
        break;
      case 4: // CmdClip
        this.ctx.beginPath();
        this.ctx.rect(cmd.x, cmd.y, cmd.w, cmd.h);
        this.ctx.clip();
        break;
      case 5: // CmdTransform
        this.ctx.setTransform(cmd.transform[0], cmd.transform[1], cmd.transform[2], cmd.transform[3], cmd.transform[4], cmd.transform[5]);
        break;
      default:
        console.warn('Unknown draw command kind:', cmd.kind);
    }
  },
  
  // Utility functions
  argbToHex: function(argb) {
    var a = (argb >> 24) & 0xFF;
    var r = (argb >> 16) & 0xFF;
    var g = (argb >> 8) & 0xFF;
    var b = argb & 0xFF;
    
    // Convert to rgba string
    return 'rgba(' + r + ',' + g + ',' + b + ',' + (a / 255) + ')';
  },
  
  // Animation control
  start: function() {
    if (this.isRunning) return;
    
    this.isRunning = true;
    this.animate();
  },
  
  stop: function() {
    this.isRunning = false;
    if (this.animationId) {
      cancelAnimationFrame(this.animationId);
      this.animationId = null;
    }
  },
  
  animate: function() {
    if (!this.isRunning) return;
    
    // Request draw commands from Go
    // This would be implemented with WebAssembly function calls
    // For now, just log
    console.log('Animation frame...');
    
    this.animationId = requestAnimationFrame(this.animate.bind(this));
  },
  
  // Canvas management
  resize: function(width, height) {
    if (this.canvas) {
      this.canvas.width = width;
      this.canvas.height = height;
    }
  },
  
  getCanvasSize: function() {
    if (this.canvas) {
      return { width: this.canvas.width, height: this.canvas.height };
    }
    return { width: 0, height: 0 };
  }
};

// Initialize on page load
document.addEventListener('DOMContentLoaded', function() {
  console.log('Aethium JS bridge loaded');
});
