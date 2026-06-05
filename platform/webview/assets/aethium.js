(function() {
    let canvas, ctx;

    const Aethium = {
        initRuntime: function(canvasID) {
            canvas = document.getElementById(canvasID);
            if (!canvas) {
                console.error('Canvas element not found:', canvasID);
                return;
            }
            ctx = canvas.getContext('2d');
            console.log('Aethium runtime initialized for canvas', canvasID);
        },

        renderFrame: function(cmds) {
            if (!ctx) return;

            // Clear canvas
            ctx.clearRect(0, 0, canvas.width, canvas.height);

            // Execute draw commands
            for (const cmd of cmds) {
                switch (cmd.Kind) {
                    case 1: // CmdFillRect
                        ctx.fillStyle = this.parseColor(cmd.Color);
                        ctx.fillRect(cmd.X, cmd.Y, cmd.W, cmd.H);
                        break;
                    case 2: // CmdStrokeRect
                        ctx.strokeStyle = this.parseColor(cmd.Color);
                        ctx.strokeRect(cmd.X, cmd.Y, cmd.W, cmd.H);
                        break;
                    case 3: // CmdDrawText
                        ctx.fillStyle = this.parseColor(cmd.Color);
                        ctx.font = '16px sans-serif';
                        ctx.fillText(cmd.Text, cmd.X, cmd.Y);
                        break;
                    case 4: // CmdClip
                        ctx.beginPath();
                        ctx.rect(cmd.X, cmd.Y, cmd.W, cmd.H);
                        ctx.clip();
                        break;
                    case 5: // CmdTransform
                        ctx.setTransform(cmd.Transform[0], cmd.Transform[1], cmd.Transform[2], cmd.Transform[3], cmd.Transform[4], cmd.Transform[5]);
                        break;
                }
            }
        },

        parseColor: function(color) {
            const a = ((color >> 24) & 0xFF) / 255;
            const r = (color >> 16) & 0xFF;
            const g = (color >> 8) & 0xFF;
            const b = color & 0xFF;
            return `rgba(${r}, ${g}, ${b}, ${a})`;
        },

        pumpEvents: function() {
            // This would be called by the Go side to process events
        }
    };

    window.Aethium = Aethium;
})();
