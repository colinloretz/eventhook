require 'spec_helper'

RSpec.describe EventHook::Sources::Stripe do
  let(:secret)  { 'whsec_test_secret' }
  let(:payload) { JSON.generate({ type: 'payment_intent.succeeded', data: { amount: 9900 } }) }
  let(:subject) { described_class.new(secret) }

  def stripe_signature(payload, secret, timestamp: Time.now.to_i)
    signed = "#{timestamp}.#{payload}"
    sig    = OpenSSL::HMAC.hexdigest('SHA256', secret, signed)
    "t=#{timestamp},v1=#{sig}"
  end

  def mock_request(body, signature, event_type: 'payment_intent.succeeded')
    double('request',
      raw_post: body,
      headers: {
        'HTTP_STRIPE_SIGNATURE' => signature,
        'Stripe-Signature'      => signature
      }
    )
  end

  it 'accepts a valid signature' do
    sig     = stripe_signature(payload, secret)
    request = mock_request(payload, sig)
    expect(subject.verify!(request)).to eq('payment_intent.succeeded')
  end

  it 'rejects a wrong secret' do
    sig     = stripe_signature(payload, 'wrong_secret')
    request = mock_request(payload, sig)
    expect { subject.verify!(request) }.to raise_error(EventHook::SignatureVerificationError, /mismatch/)
  end

  it 'rejects a stale timestamp' do
    old_timestamp = Time.now.to_i - 400
    sig     = stripe_signature(payload, secret, timestamp: old_timestamp)
    request = mock_request(payload, sig)
    expect { subject.verify!(request) }.to raise_error(EventHook::SignatureVerificationError, /too old/)
  end

  it 'rejects a missing header' do
    request = double('request', raw_post: payload, headers: {})
    expect { subject.verify!(request) }.to raise_error(EventHook::SignatureVerificationError, /Missing/)
  end
end
