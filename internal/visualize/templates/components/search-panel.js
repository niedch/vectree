class SearchPanel extends HTMLElement {
  constructor() {
    super();
    this.attachShadow({ mode: 'open' });
    this.shadowRoot.innerHTML = `
      <style>
        :host { display: none; }
        :host(.visible) { display: block; }
        .controls {
          position: absolute;
          top: 20px;
          left: 20px;
          background: rgba(255, 255, 255, 0.95);
          padding: 15px;
          border-radius: 8px;
          box-shadow: 0 4px 15px rgba(0,0,0,0.2);
          z-index: 100;
          display: flex;
          flex-direction: column;
          gap: 10px;
          width: 350px;
        }
        .input-row {
          display: flex;
          gap: 10px;
        }
        input {
          padding: 8px 12px;
          border: 1px solid #ccc;
          border-radius: 4px;
          flex-grow: 1;
          font-size: 14px;
        }
        button {
          padding: 8px 16px;
          background: #007bff;
          color: white;
          border: none;
          border-radius: 4px;
          cursor: pointer;
          font-size: 14px;
          white-space: nowrap;
        }
        button:hover { background: #0056b3; }
        button:disabled { background: #ccc; cursor: not-allowed; }
        #results {
          margin-top: 15px;
          border-top: 1px solid #eee;
          padding-top: 15px;
          max-height: 70vh;
          overflow-y: auto;
        }
        .result-item {
          background: #f8f9fa;
          border: 1px solid #e9ecef;
          border-radius: 4px;
          padding: 10px;
          margin-bottom: 10px;
          font-size: 13px;
          line-height: 1.4;
          cursor: pointer;
        }
        .result-item:hover { background: #e9ecef; }
        .result-header {
          font-weight: bold;
          color: #007bff;
          margin-bottom: 5px;
          display: flex;
          justify-content: space-between;
        }
        .result-text h1, .result-text h2, .result-text h3 { font-size: 1.1em; margin: 5px 0; }
        .result-text p { margin: 5px 0; }
        .result-text pre { background: #eee; padding: 5px; border-radius: 3px; overflow-x: auto; }
        .result-text code { background: #eee; padding: 1px 3px; border-radius: 2px; }
      </style>
      <div class="controls">
        <div class="input-row">
          <input type="text" id="promptInput" placeholder="Enter a prompt...">
          <button id="searchBtn">Project</button>
        </div>
        <div id="results" style="display:none;">
          <h4 style="margin:0 0 10px 0;">Nearest Documents</h4>
          <div id="resultsList"></div>
        </div>
      </div>
    `;
    this.shadowRoot.querySelector('#searchBtn').addEventListener('click', () => this._onSubmit());
    this.shadowRoot.querySelector('#promptInput').addEventListener('keypress', (e) => {
      if (e.key === 'Enter') this._onSubmit();
    });
  }

  _onSubmit() {
    const input = this.shadowRoot.querySelector('#promptInput');
    const prompt = input.value.trim();
    if (!prompt) return;
    this.dispatchEvent(new CustomEvent('search-panel-project', { detail: { prompt } }));
  }

  setLoading(loading) {
    const btn = this.shadowRoot.querySelector('#searchBtn');
    const input = this.shadowRoot.querySelector('#promptInput');
    btn.disabled = loading;
    btn.textContent = loading ? 'Searching...' : 'Project';
    input.disabled = loading;
  }

  setResults(docs) {
    const results = this.shadowRoot.querySelector('#results');
    const list = this.shadowRoot.querySelector('#resultsList');
    list.innerHTML = '';
    if (!docs || docs.length === 0) {
      results.style.display = 'none';
      return;
    }
    docs.forEach((doc, idx) => {
      const item = document.createElement('div');
      item.className = 'result-item';
      item.innerHTML = `
        <div class="result-header">
          <span>#${idx + 1} Document ID: ${doc.id}</span>
        </div>
        <div class="result-text">${marked.parse(doc.text)}</div>
      `;
      item.addEventListener('click', () => {
        this.dispatchEvent(new CustomEvent('search-panel-doc-select', { detail: { id: doc.id } }));
      });
      list.appendChild(item);
    });
    results.style.display = 'block';
  }
}

customElements.define('search-panel', SearchPanel);
