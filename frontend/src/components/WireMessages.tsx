import React, { useEffect, useState } from "react";
import { useNavigate } from "react-router-dom";

// Define the type for a wire message
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

const ITEMS_PER_PAGE = 5;

const WireMessages = () => {
  const [messages, setMessages] = useState<WireMessage[]>([]);
  const [error, setError] = useState("");
  const [currentPage, setCurrentPage] = useState(1);
  const [hasMore, setHasMore] = useState(false);
  const navigate = useNavigate();

  const [newMessage, setNewMessage] = useState({
    seq: "",
    sender_rtn: "",
    sender_an: "",
    receiver_rtn: "",
    receiver_an: "",
    amount: "",
  });

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    const messageString = `seq=${newMessage.seq};sender_rtn=${newMessage.sender_rtn};sender_an=${newMessage.sender_an};receiver_rtn=${newMessage.receiver_rtn};receiver_an=${newMessage.receiver_an};amount=${newMessage.amount}`;
    console.log("Sending message:", messageString);

    try {
      const response = await fetch("http://localhost:8080/wire-messages", {
        method: "POST",
        credentials: "include",
        headers: {
          "Content-Type": "text/plain",
        },
        body: messageString,
      });

      if (!response.ok) {
        const errorData = await response.json();
        console.error("Error response:", errorData);
        setError(errorData.error || "Failed to submit message");
      } else {
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
      console.error("Fetch error:", error);
      setError("Error submitting message");
    }
  };

  const handleChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    setNewMessage({
      ...newMessage,
      [e.target.name]: e.target.value,
    });
  };

  const fetchMessages = () => {
    fetch(
      `http://localhost:8080/wire-messages?page=${currentPage}&limit=${ITEMS_PER_PAGE}`,
      {
        credentials: "include",
      }
    )
      .then((response) => {
        if (response.status === 401) {
          navigate("/login");
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
        console.error("Error:", err);
        setError("Failed to fetch messages");
      });
  };

  const handleNextPage = () => {
    setCurrentPage((prev) => prev + 1);
  };

  const handlePrevPage = () => {
    setCurrentPage((prev) => Math.max(1, prev - 1));
  };

  useEffect(() => {
    fetchMessages();
  }, [currentPage, navigate]);

  return (
    <div>
      <h2>Wire Messages</h2>
      {error && <div className="error">{error}</div>}
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
      <div className="pagination">
        {currentPage > 1 && <button onClick={handlePrevPage}>Previous</button>}
        {hasMore && <button onClick={handleNextPage}>Next</button>}
      </div>
    </div>
  );
};

export default WireMessages;
