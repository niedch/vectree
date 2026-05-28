class EmbeddingPlot extends HTMLElement {
  constructor() {
    super();
    this._plotData = [];
    this._promptTraceIndex = -1;
    this._connectionsTraceIndex = -1;
  }

  connectedCallback() {
    this.style.display = 'block';
    this.style.width = '100vw';
    this.style.height = '100vh';
  }

  _getPlotEl() {
    return this;
  }

  setMainTrace(x, y, z, text, ids) {
    const trace = {
      x, y, z,
      customdata: ids,
      mode: 'markers',
      type: 'scatter3d',
      name: 'Documents',
      text,
      hoverinfo: 'text',
      marker: {
        size: 3,
        color: z,
        colorscale: 'Viridis',
        opacity: 0.6
      }
    };
    this._plotData = [trace];
    const layout = {
      title: '3D Embedding Visualization (PCA)',
      margin: { l: 0, r: 0, b: 0, t: 40 },
      legend: { x: 0, y: 1 },
      scene: {
        xaxis: { title: 'PC1' },
        yaxis: { title: 'PC2' },
        zaxis: { title: 'PC3' }
      }
    };
    Plotly.newPlot(this._getPlotEl(), this._plotData, layout);
    this._getPlotEl().on('plotly_click', (data) => {
      const point = data.points[0];
      if (point.customdata) {
        this.dispatchEvent(new CustomEvent('plot-point-click', { detail: { id: point.customdata } }));
      }
    });
  }

  updatePromptTrace(data) {
    const promptTrace = {
      x: [data.x],
      y: [data.y],
      z: [data.z],
      mode: 'markers+text',
      type: 'scatter3d',
      name: 'User Prompt',
      text: [data.text],
      textposition: 'top center',
      marker: {
        size: 8,
        color: 'red',
        symbol: 'diamond',
        line: { color: 'white', width: 2 }
      }
    };

    const connectionsTrace = { x: [], y: [], z: [], mode: 'lines', type: 'scatter3d', name: 'Connections', line: { color: 'rgba(255, 0, 0, 0.4)', width: 2 }, hoverinfo: 'none' };
    if (data.nearest_documents) {
      data.nearest_documents.forEach(doc => {
        connectionsTrace.x.push(data.x, doc.x, null);
        connectionsTrace.y.push(data.y, doc.y, null);
        connectionsTrace.z.push(data.z, doc.z, null);
      });
    }

    if (this._promptTraceIndex === -1) {
      this._plotData.push(promptTrace);
      this._promptTraceIndex = this._plotData.length - 1;
      this._plotData.push(connectionsTrace);
      this._connectionsTraceIndex = this._plotData.length - 1;
      Plotly.redraw(this._getPlotEl());
    } else {
      this._plotData[this._promptTraceIndex] = promptTrace;
      this._plotData[this._connectionsTraceIndex] = connectionsTrace;
      Plotly.react(this._getPlotEl(), this._plotData, this._getPlotEl().layout);
    }
  }

  focusCamera(x, y, z) {
    Plotly.relayout(this._getPlotEl(), {
      'scene.camera': {
        center: { x: 0, y: 0, z: 0 },
        eye: { x: (x || 1.5) * 2, y: (y || 1.5) * 2, z: (z || 1.5) * 2 }
      }
    });
  }
}

customElements.define('embedding-plot', EmbeddingPlot);
