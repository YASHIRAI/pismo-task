import streamlit as st
import requests
import json
import pandas as pd
from datetime import datetime

st.set_page_config(
    page_title="Pismo Financial Services",
    page_icon="🏦",
    layout="wide",
    initial_sidebar_state="expanded"
)

st.set_option('deprecation.showPyplotGlobalUse', False)
API_BASE_URL = "http://localhost:8083"

# Constants for UI text
ACCOUNT_ID_LABEL = "Account ID"
ACCOUNT_ID_PLACEHOLDER = "Enter account ID"
ACCOUNT_ID_REQUIRED = "Account ID is required"
METRIC_CARD_START = '<div class="metric-card">'
METRIC_CARD_END = '</div>' 

st.markdown("""
<style>
    .main-header {
        font-size: 2.5rem;
        font-weight: bold;
        color: #1f77b4;
        text-align: center;
        margin-bottom: 2rem;
    }
    .section-header {
        font-size: 1.5rem;
        font-weight: bold;
        color: #2c3e50;
        margin-top: 2rem;
        margin-bottom: 1rem;
        border-bottom: 2px solid #3498db;
        padding-bottom: 0.5rem;
    }
    .success-message {
        background-color: #d4edda;
        color: #155724;
        padding: 1rem;
        border-radius: 0.5rem;
        border: 1px solid #c3e6cb;
        margin: 1rem 0;
    }
    .error-message {
        background-color: #f8d7da;
        color: #721c24;
        padding: 1rem;
        border-radius: 0.5rem;
        border: 1px solid #f5c6cb;
        margin: 1rem 0;
    }
    .info-box {
        background-color: #d1ecf1;
        color: #0c5460;
        padding: 1rem;
        border-radius: 0.5rem;
        border: 1px solid #bee5eb;
        margin: 1rem 0;
    }
    .metric-card {
        background-color: #f8f9fa;
        padding: 1rem;
        border-radius: 0.5rem;
        border: 1px solid #dee2e6;
        text-align: center;
        margin: 0.5rem 0;
    }
</style>
""", unsafe_allow_html=True)

def make_api_request(method, endpoint, data=None, params=None):
    """Make API request to the gateway service"""
    try:
        url = f"{API_BASE_URL}{endpoint}"
        headers = {"Content-Type": "application/json"}
        
        if method.upper() == "GET":
            response = requests.get(url, params=params, timeout=10)
        elif method.upper() == "POST":
            response = requests.post(url, json=data, headers=headers, timeout=10)
        else:
            return None, f"Unsupported method: {method}"
        
        if response.status_code == 200:
            return response.json(), None
        else:
            try:
                error_data = response.json()
                return None, error_data.get("error", f"HTTP {response.status_code}: {response.text}")
            except:
                return None, f"HTTP {response.status_code}: {response.text}"
                
    except requests.exceptions.ConnectionError:
        return None, "Connection failed. Please ensure the gateway service is running on localhost:8083"
    except requests.exceptions.Timeout:
        return None, "Request timeout. Please try again."
    except Exception as e:
        return None, f"Unexpected error: {str(e)}"

def check_health():
    """Check system health"""
    data, error = make_api_request("GET", "/health")
    if error:
        return False, error
    return True, data

def display_success_message(message):
    """Display success message with styling"""
    st.markdown(f'<div class="success-message">{message}</div>', unsafe_allow_html=True)

def display_error_message(message):
    """Display error message with styling"""
    st.markdown(f'<div class="error-message">Error: {message}</div>', unsafe_allow_html=True)

def display_info_message(message):
    """Display info message with styling"""
    st.markdown(f'<div class="info-box">{message}</div>', unsafe_allow_html=True)

def create_account_tab():
    """Create Account Tab"""
    st.markdown("### Create New Account")
    with st.form("create_account_form"):
        col1, col2 = st.columns(2)
        
        with col1:
            document_number = st.text_input("Document Number", placeholder="Enter document number")
            account_type = st.selectbox("Account Type", ["CHECKING", "SAVINGS", "CREDIT"])
        
        with col2:
            initial_balance = st.number_input("Initial Balance", min_value=0.0, value=0.0, step=0.01, format="%.2f")
        
        submitted = st.form_submit_button("Create Account", use_container_width=True)
        
        if submitted:
            if not document_number:
                st.error("Document number is required")
            else:
                data = {
                    "document_number": document_number,
                    "account_type": account_type,
                    "initial_balance": initial_balance
                }
                
                result, error = make_api_request("POST", "/accounts", data)
                if error:
                    display_error_message(error)
                else:
                    display_success_message("Account created successfully!")
                    st.json(result)

def view_account_tab():
    """View Account Tab"""
    st.markdown("### View Account Details")
    with st.form("view_account_form"):
        account_id = st.text_input(ACCOUNT_ID_LABEL, placeholder=ACCOUNT_ID_PLACEHOLDER)
        submitted = st.form_submit_button("Get Account", use_container_width=True)
        
        if submitted:
            if not account_id:
                st.error(ACCOUNT_ID_REQUIRED)
            else:
                result, error = make_api_request("GET", f"/accounts/{account_id}")
                if error:
                    display_error_message(error)
                else:
                    display_success_message("Account found!")
                    
                    col1, col2, col3 = st.columns(3)
                    with col1:
                        st.metric(ACCOUNT_ID_LABEL, result["id"])
                    with col2:
                        st.metric("Account Type", result["account_type"])
                    with col3:
                        st.metric("Balance", f"${result['balance']:.2f}")
                    
                    st.json(result)

def check_balance_tab():
    """Check Balance Tab"""
    st.markdown("### Check Account Balance")
    with st.form("balance_form"):
        account_id = st.text_input(ACCOUNT_ID_LABEL, placeholder=ACCOUNT_ID_PLACEHOLDER, key="balance_account_id")
        submitted = st.form_submit_button("Get Balance", use_container_width=True)
        
        if submitted:
            if not account_id:
                st.error(ACCOUNT_ID_REQUIRED)
            else:
                result, error = make_api_request("GET", f"/accounts/{account_id}/balance")
                if error:
                    display_error_message(error)
                else:
                    display_success_message("Balance retrieved successfully!")
                    st.metric("Current Balance", f"${result['balance']:.2f}")

def create_account_ui():
    """Account Management UI"""
    st.markdown('<div class="section-header">Account Management</div>', unsafe_allow_html=True)
    
    tab1, tab2, tab3 = st.tabs(["Create Account", "View Account", "Check Balance"])
    
    with tab1:
        create_account_tab()
    
    with tab2:
        view_account_tab()
    
    with tab3:
        check_balance_tab()

def create_transaction_tab():
    """Create Transaction Tab"""
    st.markdown("### Create New Transaction")
    with st.form("create_transaction_form"):
        col1, col2 = st.columns(2)
        
        with col1:
            account_id = st.text_input(ACCOUNT_ID_LABEL, placeholder=ACCOUNT_ID_PLACEHOLDER)
            operation_type = st.selectbox("Operation Type", [
                "CASH_PURCHASE",
                "INSTALLMENT_PURCHASE", 
                "WITHDRAWAL",
                "PAYMENT"
            ])
        
        with col2:
            amount = st.number_input("Amount", min_value=0.01, value=0.01, step=0.01, format="%.2f")
            description = st.text_area("Description", placeholder="Enter transaction description")
        
        submitted = st.form_submit_button("Create Transaction", use_container_width=True)
        
        if submitted:
            if not account_id:
                st.error(ACCOUNT_ID_REQUIRED)
            else:
                data = {
                    "account_id": account_id,
                    "operation_type": operation_type,
                    "amount": amount,
                    "description": description
                }
                
                result, error = make_api_request("POST", "/transactions", data)
                if error:
                    st.markdown(f'<div class="error-message">Error: {error}</div>', unsafe_allow_html=True)
                else:
                    st.markdown('<div class="success-message">Transaction created successfully!</div>', unsafe_allow_html=True)
                    st.json(result)

def view_transaction_tab():
    """View Transaction Tab"""
    st.markdown("### View Transaction Details")
    with st.form("view_transaction_form"):
        transaction_id = st.text_input("Transaction ID", placeholder="Enter transaction ID")
        submitted = st.form_submit_button("Get Transaction", use_container_width=True)
        
        if submitted:
            if not transaction_id:
                st.error("Transaction ID is required")
            else:
                result, error = make_api_request("GET", f"/transactions/{transaction_id}")
                if error:
                    st.markdown(f'<div class="error-message">Error: {error}</div>', unsafe_allow_html=True)
                else:
                    st.markdown('<div class="success-message">Transaction found!</div>', unsafe_allow_html=True)
                    
                    col1, col2, col3 = st.columns(3)
                    with col1:
                        st.metric("Transaction ID", result["id"])
                    with col2:
                        st.metric("Operation Type", result["operation_type"])
                    with col3:
                        st.metric("Amount", f"${result['amount']:.2f}")
                    
                    col4, col5 = st.columns(2)
                    with col4:
                        st.metric("Status", result["status"])
                    with col5:
                        st.metric(ACCOUNT_ID_LABEL, result["account_id"])
                    
                    st.json(result)

def transaction_history_tab():
    """Transaction History Tab"""
    st.markdown("### Transaction History")
    with st.form("history_form"):
        account_id = st.text_input(ACCOUNT_ID_LABEL, placeholder=ACCOUNT_ID_PLACEHOLDER, key="history_account_id")
        
        col1, col2 = st.columns(2)
        with col1:
            limit = st.number_input("Limit", min_value=1, max_value=100, value=20)
        with col2:
            offset = st.number_input("Offset", min_value=0, value=0)
        
        submitted = st.form_submit_button("Get History", use_container_width=True)
        
        if submitted:
            if not account_id:
                st.error(ACCOUNT_ID_REQUIRED)
            else:
                params = {"limit": limit, "offset": offset}
                result, error = make_api_request("GET", f"/accounts/{account_id}/transactions", params=params)
                if error:
                    st.markdown(f'<div class="error-message">Error: {error}</div>', unsafe_allow_html=True)
                else:
                    st.markdown('<div class="success-message">Transaction history retrieved!</div>', unsafe_allow_html=True)
                    
                    st.metric("Total Transactions", result["total"])
                    
                    if result["transactions"]:
                        df = pd.DataFrame(result["transactions"])
                        df['created_at'] = pd.to_datetime(df['created_at'], unit='s')
                        df['amount'] = df['amount'].apply(lambda x: f"${x:.2f}")
                        
                        st.dataframe(
                            df[['id', 'operation_type', 'amount', 'status', 'description', 'created_at']],
                            use_container_width=True
                        )
                    else:
                        st.info("No transactions found for this account.")

def process_payment_tab():
    """Process Payment Tab"""
    st.markdown("### Process Payment")
    with st.form("payment_form"):
        col1, col2 = st.columns(2)
        
        with col1:
            account_id = st.text_input(ACCOUNT_ID_LABEL, placeholder=ACCOUNT_ID_PLACEHOLDER, key="payment_account_id")
            amount = st.number_input("Payment Amount", min_value=0.01, value=0.01, step=0.01, format="%.2f")
        
        with col2:
            description = st.text_area("Payment Description", placeholder="Enter payment description", key="payment_description")
        
        submitted = st.form_submit_button("Process Payment", use_container_width=True)
        
        if submitted:
            if not account_id:
                st.error(ACCOUNT_ID_REQUIRED)
            else:
                data = {
                    "account_id": account_id,
                    "amount": amount,
                    "description": description
                }
                
                result, error = make_api_request("POST", "/payments", data)
                if error:
                    st.markdown(f'<div class="error-message">Error: {error}</div>', unsafe_allow_html=True)
                else:
                    st.markdown('<div class="success-message">Payment processed successfully!</div>', unsafe_allow_html=True)
                    st.json(result)

def create_transaction_ui():
    """Transaction Management UI"""
    st.markdown('<div class="section-header">Transaction Management</div>', unsafe_allow_html=True)
    
    tab1, tab2, tab3, tab4 = st.tabs(["Create Transaction", "View Transaction", "Transaction History", "Process Payment"])
    
    with tab1:
        create_transaction_tab()
    
    with tab2:
        view_transaction_tab()
    
    with tab3:
        transaction_history_tab()
    
    with tab4:
        process_payment_tab()

def health_monitoring_tab():
    """Health Monitoring Tab"""
    st.markdown('<div class="section-header">System Health</div>', unsafe_allow_html=True)
    
    col1, col2 = st.columns([1, 1])
    
    with col1:
        if st.button("Check System Health", use_container_width=True):
            with st.spinner("Checking system health..."):
                is_healthy, data = check_health()
                
                if is_healthy:
                    st.markdown('<div class="success-message">System is healthy!</div>', unsafe_allow_html=True)
                    st.json(data)
                else:
                    st.markdown(f'<div class="error-message">System error: {data}</div>', unsafe_allow_html=True)
    
    with col2:
        st.markdown("### Service Status")
        st.markdown("""
        - **Gateway Service**: Port 8083
        - **Account Manager**: Port 8081  
        - **Transaction Manager**: Port 8082
        - **Database**: PostgreSQL
        """)

def dashboard_tab():
    """Dashboard Tab"""
    st.markdown('<div class="section-header">System Dashboard</div>', unsafe_allow_html=True)
    
    col1, col2, col3 = st.columns(3)
    
    with col1:
        st.markdown(METRIC_CARD_START, unsafe_allow_html=True)
        st.metric("Gateway Status", "Online", "Port 8083")
        st.markdown(METRIC_CARD_END, unsafe_allow_html=True)
    
    with col2:
        st.markdown(METRIC_CARD_START, unsafe_allow_html=True)
        st.metric("Account Service", "Online", "Port 8081")
        st.markdown(METRIC_CARD_END, unsafe_allow_html=True)
    
    with col3:
        st.markdown(METRIC_CARD_START, unsafe_allow_html=True)
        st.metric("Transaction Service", "Online", "Port 8082")
        st.markdown(METRIC_CARD_END, unsafe_allow_html=True)
    
    st.markdown("### Recent Activity")
    st.info("Use the Account and Transaction tabs to view and manage data.")
    
    if st.button("Refresh Health Status", use_container_width=True):
        is_healthy, data = check_health()
        if is_healthy:
            st.success("System is healthy")
        else:
            st.error(f"System error: {data}")

def create_health_monitoring():
    """Health Monitoring UI - Legacy function for backward compatibility"""
    health_monitoring_tab()

def main():
    """Main application"""
    st.markdown('<div class="main-header">Pismo Financial Services</div>', unsafe_allow_html=True)
    
    with st.sidebar:
        st.markdown("### Quick Actions")
        
        if st.button("Check Health", use_container_width=True):
            with st.spinner("Checking..."):
                is_healthy, data = check_health()
                if is_healthy:
                    st.success("System Healthy")
                else:
                    st.error(f"Error: {data}")
        
        st.markdown("### System Info")
        st.info("""
        **API Gateway**: localhost:8083
        
        **Services**:
        - Account Manager (8081)
        - Transaction Manager (8082)
        
        **Database**: PostgreSQL
        """)
        
        st.markdown("### Configuration")
        api_url = st.text_input("API Base URL", value=API_BASE_URL)
        if api_url != API_BASE_URL:
            st.warning("Restart app to apply new API URL")
    
    tab1, tab2, tab3, tab4 = st.tabs(["Accounts", "Transactions", "Dashboard", "Health"])
    
    with tab1:
        create_account_ui()
    
    with tab2:
        create_transaction_ui()
    
    with tab3:
        dashboard_tab()
    
    with tab4:
        health_monitoring_tab()

if __name__ == "__main__":
    main()
