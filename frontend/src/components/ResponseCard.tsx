import type { CompletionResponse } from '../types/api';

type ResponseCardProps = {
  response: CompletionResponse;
};

export function ResponseCard({ response }: ResponseCardProps) {
  const { gateway_meta, choices, usage } = response;
  
  return (
    <div className="info-tile">
      <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', marginBottom: '8px' }}>
        <h3 style={{ display: 'flex', alignItems: 'center', gap: '12px', margin: 0 }}>
          <span style={{ textTransform: 'capitalize' }}>
            Provider: {gateway_meta.provider}
          </span>
          {gateway_meta.fallback_used && (
            <span style={{ 
              fontSize: '0.75rem', 
              padding: '4px 10px', 
              borderRadius: '999px', 
              background: '#ffe4e6', 
              color: '#e11d48', 
              fontWeight: 700,
              letterSpacing: '0.05em',
              textTransform: 'uppercase'
            }}>
              Fallback
            </span>
          )}
        </h3>
      </div>
      
      <div className="mode-panel__details" style={{ marginBottom: '16px', fontSize: '0.9rem', color: 'rgba(70, 54, 42, 0.82)', paddingBottom: '16px', borderBottom: '1px solid rgba(68, 48, 32, 0.08)' }}>
        <div style={{ display: 'flex', flexDirection: 'column', gap: '4px' }}>
          <span><strong>Latency:</strong> {gateway_meta.latency_ms} ms</span>
          <span><strong>Cost:</strong> ${gateway_meta.cost_estimate_usd.toFixed(6)}</span>
        </div>
        <div style={{ display: 'flex', flexDirection: 'column', gap: '4px' }}>
          <span><strong>Tokens:</strong> {usage.total_tokens}</span>
        </div>
      </div>
      
      <div style={{ 
        background: 'rgba(255, 255, 255, 0.6)', 
        padding: '20px', 
        borderRadius: '16px', 
        whiteSpace: 'pre-wrap',
        color: '#22160d',
        fontSize: '1rem',
        lineHeight: '1.6',
        border: '1px solid rgba(68, 48, 32, 0.04)'
      }}>
        {choices[0]?.message?.content}
      </div>
    </div>
  );
}
