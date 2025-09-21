// MongoDB initialization script for Pismo microservices

db = db.getSiblingDB('accounts');

db.createCollection('accounts', {
  validator: {
    $jsonSchema: {
      bsonType: 'object',
      required: ['_id', 'document_number', 'account_type', 'balance', 'created_at', 'updated_at'],
      properties: {
        _id: {
          bsonType: 'string',
          description: 'Account ID must be a string and is required'
        },
        document_number: {
          bsonType: 'string',
          description: 'Document number must be a string and is required'
        },
        account_type: {
          bsonType: 'string',
          enum: ['CHECKING', 'SAVINGS', 'CREDIT'],
          description: 'Account type must be one of: CHECKING, SAVINGS, CREDIT'
        },
        balance: {
          bsonType: 'double',
          minimum: 0,
          description: 'Balance must be a double and cannot be negative'
        },
        created_at: {
          bsonType: 'date',
          description: 'Created at must be a date and is required'
        },
        updated_at: {
          bsonType: 'date',
          description: 'Updated at must be a date and is required'
        }
      }
    }
  }
});

db.accounts.createIndex({ 'document_number': 1 }, { unique: true });
db.accounts.createIndex({ 'account_type': 1 });
db.accounts.createIndex({ 'created_at': -1 });

db = db.getSiblingDB('transactions');

db.createCollection('transactions', {
  validator: {
    $jsonSchema: {
      bsonType: 'object',
      required: ['_id', 'account_id', 'operation_type', 'amount', 'created_at', 'status'],
      properties: {
        _id: {
          bsonType: 'string',
          description: 'Transaction ID must be a string and is required'
        },
        account_id: {
          bsonType: 'string',
          description: 'Account ID must be a string and is required'
        },
        operation_type: {
          bsonType: 'string',
          enum: ['COMPRA_A_VISTA', 'COMPRA_PARCELADA', 'SAQUE', 'PAGAMENTO'],
          description: 'Operation type must be one of: COMPRA_A_VISTA, COMPRA_PARCELADA, SAQUE, PAGAMENTO'
        },
        amount: {
          bsonType: 'double',
          description: 'Amount must be a double and is required'
        },
        description: {
          bsonType: 'string',
          description: 'Description must be a string'
        },
        created_at: {
          bsonType: 'date',
          description: 'Created at must be a date and is required'
        },
        status: {
          bsonType: 'string',
          enum: ['PENDING', 'COMPLETED', 'FAILED', 'CANCELLED'],
          description: 'Status must be one of: PENDING, COMPLETED, FAILED, CANCELLED'
        }
      }
    }
  }
});

db.transactions.createIndex({ 'account_id': 1 });
db.transactions.createIndex({ 'created_at': -1 });
db.transactions.createIndex({ 'account_id': 1, 'created_at': -1 });
db.transactions.createIndex({ 'operation_type': 1 });
db.transactions.createIndex({ 'status': 1 });

print('MongoDB initialization completed successfully!');
print('Created databases: accounts, transactions');
print('Created collections with validation and indexes');
