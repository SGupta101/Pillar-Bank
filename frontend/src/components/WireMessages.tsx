import React, { useEffect, useState } from "react";
import { useNavigate } from "react-router-dom";

const WireMessages = () => {
  const [messages, setMessages] = useState([]);
  const [error, setError] = useState("");
  const navigate = useNavigate();

  useEffect(() => {
    fetch("http://localhost:8080/wire-messages", {
      credentials: "include", // Important: This sends the cookie
    })
      .then((response) => {
        if (response.status === 401) {
          // Unauthorized, token invalid or expired
          navigate("/login");
          return;
        }
        return response.json();
      })
      .then((data) => {
        if (data) setMessages(data);
      })
      .catch((err) => {
        setError("Failed to fetch messages");
      });
  }, [navigate]);

  return (
    <div>
      <h2>Wire Messages</h2>
      {error && <div className="error">{error}</div>}
      {/* Display messages here */}
    </div>
  );
};

export default WireMessages;
