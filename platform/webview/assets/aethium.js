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
            ctx.save(); // Save initial state
            console.log('Aethium runtime initialized for canvas', canvasID);
        },

        renderFrame: function(cmds) {
            if (!ctx) return;

            // Reset state (including clipping)
            ctx.restore();
            ctx.save();

            // Clear canvas
            ctx.clearRect(0, 0, canvas.width, canvas.height);

            // Execute draw commands
            for (const cmd of cmds) {
                const kind = cmd[0];
                switch (kind) {
                    case 1: // CmdFillRect
                        ctx.fillStyle = this.parseColor(cmd[5]);
                        ctx.fillRect(cmd[1], cmd[2], cmd[3], cmd[4]);
                        break;
                    case 2: // CmdStrokeRect
                        ctx.strokeStyle = this.parseColor(cmd[5]);
                        ctx.strokeRect(cmd[1], cmd[2], cmd[3], cmd[4]);
                        break;
                    case 3: // CmdDrawText
                        ctx.fillStyle = this.parseColor(cmd[5]);
                        ctx.font = '16px sans-serif';
                        ctx.fillText(cmd[6], cmd[1], cmd[2]);
                        break;
                    case 4: // CmdClip
                        ctx.beginPath();
                        ctx.rect(cmd[1], cmd[2], cmd[3], cmd[4]);
                        ctx.clip();
                        break;
                    case 5: // CmdTransform
                        const m = cmd[7];
                        ctx.setTransform(m[0], m[1], m[2], m[3], m[4], m[5]);
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
