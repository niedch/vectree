class DocOverlay extends HTMLElement {
  constructor() {
    super();
    this.attachShadow({ mode: 'open' });
    this.shadowRoot.innerHTML = `
      <style>
        :host { display: none; }
        :host(.open) { display: flex; }
        .overlay {
          position: absolute;
          top: 20px;
          right: 20px;
          width: 400px;
          max-height: 90vh;
          background: rgba(255, 255, 255, 0.98);
          border-radius: 8px;
          box-shadow: 0 4px 20px rgba(0,0,0,0.3);
          z-index: 200;
          flex-direction: column;
        }
        .header {
          padding: 15px;
          border-bottom: 1px solid #eee;
          display: flex;
          justify-content: space-between;
          align-items: center;
          font-weight: bold;
        }
        .close-btn {
          cursor: pointer;
          font-size: 20px;
          color: #999;
          background: none;
          border: none;
          padding: 0;
        }
        .close-btn:hover { color: #333; }
        .content {
          padding: 20px;
          overflow-y: auto;
          font-size: 14px;
          line-height: 1.6;
          color: #333;
        }
        .content h1, .content h2, .content h3 { margin-top: 10px; margin-bottom: 10px; }
        .content code { background: #f4f4f4; padding: 2px 4px; border-radius: 4px; font-family: monospace; }
        .content pre { background: #f4f4f4; padding: 10px; border-radius: 4px; overflow-x: auto; margin: 10px 0; }
        .content pre code { background: transparent; padding: 0; }
        .content img { max-width: 100%; }
        .content ul, .content ol { padding-left: 20px; }
        .parent-link {
          margin-bottom: 15px;
          padding-bottom: 10px;
          border-bottom: 1px solid #eee;
        }
        .parent-link a {
          color: #007bff;
          text-decoration: none;
          font-size: 13px;
          cursor: pointer;
        }
      </style>
      <div class="overlay">
        <div class="header">
          <span id="title">Document Content</span>
          <button class="close-btn">&times;</button>
        </div>
        <div class="content">
          <div id="parentLink" class="parent-link"></div>
          <div id="body"></div>
        </div>
      </div>
    `;
    this.shadowRoot.querySelector('.close-btn').addEventListener('click', () => {
      this.dispatchEvent(new CustomEvent('doc-overlay-close'));
    });
  }

  show(id, content, parentId) {
    this.classList.add('open');
    this.shadowRoot.querySelector('#title').textContent = `Document ID: ${id}`;
    const linkEl = this.shadowRoot.querySelector('#parentLink');
    if (parentId) {
      linkEl.innerHTML = '';
      const a = document.createElement('a');
      a.textContent = `View Parent Section (ID: ${parentId})`;
      a.addEventListener('click', (e) => {
        e.preventDefault();
        this.dispatchEvent(new CustomEvent('doc-overlay-navigate-parent', { detail: { id: parentId } }));
      });
      linkEl.appendChild(a);
    } else {
      linkEl.innerHTML = '';
    }
    this.shadowRoot.querySelector('#body').innerHTML = content;
  }

  hide() {
    this.classList.remove('open');
  }
}

customElements.define('doc-overlay', DocOverlay);
