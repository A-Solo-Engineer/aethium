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
                        this.fillRect(cmd[1], cmd[2], cmd[3], cmd[4], cmd[5]);
                        break;
                    case 2: // CmdStrokeRect
                        this.strokeRect(cmd[1], cmd[2], cmd[3], cmd[4], cmd[5]);
                        break;
                    case 3: // CmdDrawText
                        this.drawText(cmd[1], cmd[2], cmd[6], cmd[5]);
                        break;
                    case 4: // CmdClip
                        this.setClip(cmd[1], cmd[2], cmd[3], cmd[4]);
                        break;
                    case 5: // CmdTransform
                        this.setTransform(cmd[7]);
                        break;
                }
            }
        },

        fillRect: function(x, y, w, h, color) {
            if (!ctx) return;
            ctx.fillStyle = this.parseColor(color);
            ctx.fillRect(x, y, w, h);
        },

        strokeRect: function(x, y, w, h, color) {
            if (!ctx) return;
            ctx.strokeStyle = this.parseColor(color);
            ctx.strokeRect(x, y, w, h);
        },

        drawText: function(x, y, text, color) {
            if (!ctx) return;
            ctx.fillStyle = this.parseColor(color);
            ctx.font = '16px sans-serif';
            ctx.fillText(text, x, y);
        },

        setClip: function(x, y, w, h) {
            if (!ctx) return;
            ctx.beginPath();
            ctx.rect(x, y, w, h);
            ctx.clip();
        },

        setTransform: function(m) {
            if (!ctx) return;
            ctx.setTransform(m[0], m[1], m[2], m[3], m[4], m[5]);
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
