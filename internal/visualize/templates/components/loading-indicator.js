class LoadingIndicator extends HTMLElement {
  constructor() {
    super();
    this.attachShadow({ mode: 'open' });
    this.shadowRoot.innerHTML = `
      <style>
        :host {
          position: absolute;
          top: 0;
          left: 0;
          width: 100vw;
          height: 100vh;
          display: flex;
          align-items: center;
          justify-content: center;
          z-index: 1000;
          pointer-events: none;
        }
        :host(.hidden) { display: none; }
        .spinner {
          background: rgba(255, 255, 255, 0.95);
          padding: 1rem 2rem;
          border-radius: 8px;
          box-shadow: 0 4px 15px rgba(0,0,0,0.1);
          font-size: 1.2rem;
          color: #555;
          pointer-events: auto;
        }
      </style>
      <div class="spinner">
        <slot>Calculating PCA...</slot>
      </div>
    `;
  }

  show(msg) {
    if (msg) this.shadowRoot.querySelector('slot').textContent = msg;
    this.classList.remove('hidden');
  }

  hide() {
    this.classList.add('hidden');
  }
}

customElements.define('loading-indicator', LoadingIndicator);
