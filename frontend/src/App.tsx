import React from "react";
import { BrowserRouter, Routes, Route, Navigate } from "react-router-dom";
import Login from "./components/Login";
import WireMessages from "./components/WireMessages";
import "./App.css";

const App = () => (
  <BrowserRouter>
    <div className="App">
      <Routes>
        <Route path="/login" element={<Login />} />
        <Route path="/wire-messages" element={<WireMessages />} />
        <Route path="/" element={<Navigate to="/login" />} />
      </Routes>
    </div>
  </BrowserRouter>
);

export default App;
