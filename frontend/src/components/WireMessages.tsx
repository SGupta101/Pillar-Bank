import React, { useEffect, useState } from "react";
import { useNavigate } from "react-router-dom";

// WireMessage defines the structure of a wire transfer message
interface WireMessage {
  id: number;
  seq: number;
  sender_rtn: string;
  sender_an: string;
  receiver_rtn: string;
  receiver_an: string;
  amount: number;
  message: string;
}

// Number of wire messages to display per page
const ITEMS_PER_PAGE = 5;

// Backend API endpoint
const API_URL = "http://localhost:8080";

// WireMessages component handles displaying and creating wire messages
const WireMessages = () => {
  const [messages, setMessages] = useState<WireMessage[]>([]);
  const [error, setError] = useState("");
  const [currentPage, setCurrentPage] = useState(1);
  const [hasMore, setHasMore] = useState(false);
  const [sortColumn, setSortColumn] = useState("seq");
  const navigate = useNavigate();

  // State for new message form
  const [newMessage, setNewMessage] = useState({
    seq: "",
    sender_rtn: "",
    sender_an: "",
    receiver_rtn: "",
    receiver_an: "",
    amount: "",
  });

  // Handle new message form submission
  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    const messageString = `seq=${newMessage.seq};sender_rtn=${newMessage.sender_rtn};sender_an=${newMessage.sender_an};receiver_rtn=${newMessage.receiver_rtn};receiver_an=${newMessage.receiver_an};amount=${newMessage.amount}`;

    try {
      const response = await fetch(`${API_URL}/wire-messages`, {
        method: "POST",
        credentials: "include", // Required for cookies
        headers: {
          "Content-Type": "text/plain",
        },
        body: messageString,
      });

      if (!response.ok) {
        const errorData = await response.json();
        setError(errorData.error || "Failed to submit message");
      } else {
        // Reset form and refresh messages on success
        fetchMessages();
        setNewMessage({
          seq: "",
          sender_rtn: "",
          sender_an: "",
          receiver_rtn: "",
          receiver_an: "",
          amount: "",
        });
      }
    } catch (error) {
      setError("Error submitting message");
    }
  };

  // Handle form input changes
  const handleChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    setNewMessage({
      ...newMessage,
      [e.target.name]: e.target.value,
    });
  };

  // Fetch paginated wire messages from backend
  const fetchMessages = () => {
    fetch(
      // `${API_URL}/wire-messages?page=${currentPage}&limit=${ITEMS_PER_PAGE}`,
      `${API_URL}/wire-messages?page=${currentPage}&limit=${ITEMS_PER_PAGE}&sort=${sortColumn}`,
      {
        credentials: "include", // Required for cookies
      }
    )
      .then((response) => {
        if (response.status === 401) {
          navigate("/login"); // Redirect to login if unauthorized
          return;
        }
        return response.json();
      })
      .then((data) => {
        if (Array.isArray(data)) {
          setMessages(data);
          setHasMore(data.length === ITEMS_PER_PAGE);
        } else if (data && data.message === "No wire messages found") {
          setMessages([]);
          setHasMore(false);
        } else {
          setError("Unexpected data format");
        }
      })
      .catch((err) => {
        setError("Failed to fetch messages");
      });
  };

  // Pagination handlers
  const handleNextPage = () => {
    setCurrentPage((prev) => prev + 1);
  };

  const handlePrevPage = () => {
    setCurrentPage((prev) => Math.max(1, prev - 1));
  };

  // Fetch messages when page changes or sort column changes
  useEffect(() => {
    fetchMessages();
  }, [currentPage, navigate, sortColumn]);

  return (
    <div>
      <h2>Wire Messages</h2>
      {error && <div className="error">{error}</div>}

      {/* New message form */}
      <div className="add-message-form">
        <h3>Add New Wire Message</h3>
        <form onSubmit={handleSubmit}>
          <input
            type="number"
            name="seq"
            placeholder="Sequence Number"
            value={newMessage.seq}
            onChange={handleChange}
            required
          />
          <input
            type="text"
            name="sender_rtn"
            placeholder="Sender RTN"
            value={newMessage.sender_rtn}
            onChange={handleChange}
            required
          />
          <input
            type="text"
            name="sender_an"
            placeholder="Sender Account Number"
            value={newMessage.sender_an}
            onChange={handleChange}
            required
          />
          <input
            type="text"
            name="receiver_rtn"
            placeholder="Receiver RTN"
            value={newMessage.receiver_rtn}
            onChange={handleChange}
            required
          />
          <input
            type="text"
            name="receiver_an"
            placeholder="Receiver Account Number"
            value={newMessage.receiver_an}
            onChange={handleChange}
            required
          />
          <input
            type="number"
            name="amount"
            placeholder="Amount"
            value={newMessage.amount}
            onChange={handleChange}
            required
          />
          <button type="submit">Add Message</button>
        </form>
      </div>

      {/* Wire messages table */}
      <table>
        <thead>
          <tr>
            <th>Sequence</th>
            <th>Sender RTN</th>
            <th>Sender Account</th>
            <th>Receiver RTN</th>
            <th>Receiver Account</th>
            <th>Amount</th>
          </tr>
        </thead>
        <tbody>
          {messages.map((msg) => (
            <tr key={msg.id}>
              <td>{msg.seq}</td>
              <td>{msg.sender_rtn}</td>
              <td>{msg.sender_an}</td>
              <td>{msg.receiver_rtn}</td>
              <td>{msg.receiver_an}</td>
              <td>${msg.amount}</td>
            </tr>
          ))}
        </tbody>
      </table>

      {/* Pagination controls */}
      <div className="pagination">
        {currentPage > 1 && <button onClick={handlePrevPage}>Previous</button>}
        {hasMore && <button onClick={handleNextPage}>Next</button>}
      </div>

      {/* Column selector */}
      <div className="column-selector">
        <label>
          Sort by:
          <select
            value={sortColumn}
            onChange={(e) => {
              setSortColumn(e.target.value);
              setCurrentPage(1); // Reset to the first page when sorting changes
            }}
          >
            <option value="seq">Sequence</option>
            <option value="sender_rtn">Sender RTN</option>
            <option value="sender_an">Sender Account</option>
            <option value="receiver_rtn">Receiver RTN</option>
            <option value="receiver_an">Receiver Account</option>
            <option value="amount">Amount</option>
          </select>
        </label>
      </div>
    </div>
  );
};

export default WireMessages;
