import type { ChangeEvent, KeyboardEvent } from 'react';

type PromptInputProps = {
  value: string;
  onChange: (val: string) => void;
  onSubmit: () => void;
  disabled?: boolean;
};

export function PromptInput({ value, onChange, onSubmit, disabled }: PromptInputProps) {
  const handleKeyDown = (e: KeyboardEvent<HTMLTextAreaElement>) => {
    if (e.key === 'Enter' && !e.shiftKey) {
      e.preventDefault();
      if (!disabled && value.trim()) {
        onSubmit();
      }
    }
  };

  return (
    <div className="mode-panel" style={{ padding: '20px' }}>
      <div className="mode-panel__copy" style={{ display: 'flex', flexDirection: 'column' }}>
        <h3 className="mode-panel__label" style={{ marginBottom: '12px' }}>Prompt</h3>
        <textarea
          value={value}
          onChange={(e: ChangeEvent<HTMLTextAreaElement>) => onChange(e.target.value)}
          onKeyDown={handleKeyDown}
          disabled={disabled}
          placeholder="Enter your prompt here... (Shift+Enter for new line)"
          style={{ 
            width: '100%', 
            minHeight: '100px', 
            resize: 'vertical', 
            padding: '16px', 
            borderRadius: '16px', 
            border: '1px solid rgba(68, 48, 32, 0.16)', 
            background: 'rgba(255, 255, 255, 0.9)', 
            fontSize: '1.05rem', 
            color: '#22160d',
            fontFamily: 'inherit',
            lineHeight: '1.5'
          }}
        />
        <button 
          onClick={onSubmit} 
          disabled={disabled || !value.trim()}
          className="app-nav__link"
          style={{ 
            alignSelf: 'flex-end', 
            marginTop: '16px', 
            cursor: (disabled || !value.trim()) ? 'not-allowed' : 'pointer',
            opacity: (disabled || !value.trim()) ? 0.6 : 1
          }}
        >
          Submit
        </button>
      </div>
    </div>
  );
}
