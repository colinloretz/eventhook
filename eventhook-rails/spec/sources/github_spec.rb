require 'spec_helper'

RSpec.describe EventHook::Sources::Github do
  let(:secret)  { 'github_test_secret' }
  let(:payload) { JSON.generate({ action: 'opened', number: 1 }) }
  let(:subject) { described_class.new(secret) }

  def github_signature(payload, secret)
    'sha256=' + OpenSSL::HMAC.hexdigest('SHA256', secret, payload)
  end

  def mock_request(body, signature, event: 'pull_request')
    double('request',
      raw_post: body,
      headers: {
        'HTTP_X_HUB_SIGNATURE_256' => signature,
        'X-Hub-Signature-256'      => signature,
        'HTTP_X_GITHUB_EVENT'      => event,
        'X-GitHub-Event'           => event
      }
    )
  end

  it 'accepts a valid signature' do
    sig     = github_signature(payload, secret)
    request = mock_request(payload, sig)
    expect(subject.verify!(request)).to eq('github.pull_request')
  end

  it 'rejects a wrong secret' do
    sig     = github_signature(payload, 'bad_secret')
    request = mock_request(payload, sig)
    expect { subject.verify!(request) }.to raise_error(EventHook::SignatureVerificationError, /mismatch/)
  end

  it 'rejects a missing header' do
    request = double('request', raw_post: payload, headers: {})
    expect { subject.verify!(request) }.to raise_error(EventHook::SignatureVerificationError, /Missing/)
  end
end
