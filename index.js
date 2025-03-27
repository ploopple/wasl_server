const express = require('express');
const { Client, Environment } = require('square');
const crypto = require('crypto');
const cors = require('cors');
const dotenv = require('dotenv');
const bodyParser = require('body-parser');

// Load environment variables
dotenv.config();

// Create Express app
const app = express();

// Fix BigInt serialization issue
// This adds a custom BigInt serializer to JSON.stringify
BigInt.prototype.toJSON = function() {
  return this.toString();
};

// Middleware
app.use(cors());
app.use(bodyParser.json());

// Initialize Square client
const squareClient = new Client({
  environment: process.env.ENVIRONMENT === 'PRODUCTION' 
    ? Environment.Production 
    : Environment.Sandbox,
  accessToken: process.env.ACCESS_TOKEN,
});

// Get Square API instances
const { paymentsApi, ordersApi } = squareClient;

// Handle charge-card requests from place_order_page.dart
app.post('/charge-card', async (req, res) => {
  try {
    const { nonce, amount, currency = 'USD' } = req.body;
    
    if (!nonce || !amount) {
      return res.status(400).json({ 
        errorMessage: 'Missing required fields: nonce and amount are required' 
      });
    }

    const locationId = process.env.LOCATION_ID;
    
    // 1. Create an order
    const amountInCents = Math.round(parseFloat(amount) * 100);
    const orderRequest = {
      idempotencyKey: crypto.randomBytes(16).toString('hex'),
      order: {
        locationId,
        lineItems: [
          {
            name: 'Wasl Order',
            quantity: '1',
            basePriceMoney: {
              amount: amountInCents,
              currency
            }
          }
        ]
      }
    };

    const orderResponse = await ordersApi.createOrder(orderRequest);
    
    // 2. Create payment with the order ID and nonce
    const paymentRequest = {
      idempotencyKey: crypto.randomBytes(16).toString('hex'),
      sourceId: nonce,
      amountMoney: {
        ...orderResponse.result.order.totalMoney,
      },
      orderId: orderResponse.result.order.id,
      autocomplete: true,
      locationId,
    };

    const paymentResponse = await paymentsApi.createPayment(paymentRequest);
    
    // 3. Return successful response - safely convert to plain object first
    const safePaymentResult = JSON.parse(JSON.stringify(paymentResponse.result.payment));
    return res.status(200).json(safePaymentResult);
  } catch (error) {
    console.error('Payment Error:', error);
    
    // Handle Square API errors
    const errorMessage = error.errors?.[0]?.detail || 'Payment processing failed';
    const errorCode = error.errors?.[0]?.code || 'UNKNOWN_ERROR';
    
    return res.status(400).json({
      errorMessage,
      errorCode
    });
  }
});

// Client token endpoint for authentication (simple implementation)
app.get('/client_token', (req, res) => {
  // Generate a random token
  const token = crypto.randomBytes(32).toString('hex');
  res.status(200).send(token);
});

// Start server
const port = process.env.PORT || 3000;
app.listen(port, () => {
  console.log(`Server running on port ${port}`);
  console.log(`Environment: ${process.env.ENVIRONMENT || 'SANDBOX'}`);
});