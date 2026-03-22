#!/bin/bash

# Script to add sample problems to the database
# Usage: ./add_sample_problems.sh

BASE_URL="http://localhost:8080/api/admin"
TOKEN=""

# Get admin token
echo "Logging in as admin..."
LOGIN_RESPONSE=$(curl -s -X POST http://localhost:8080/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"admin123"}')

TOKEN=$(echo $LOGIN_RESPONSE | grep -o '"token":"[^"]*' | cut -d'"' -f4)

if [ -z "$TOKEN" ]; then
  echo "Failed to get admin token. Make sure admin user exists."
  exit 1
fi

echo "Admin token obtained successfully"
echo ""

# Problem 1: Two Sum
echo "Adding Problem 1: Two Sum..."
curl -X POST "$BASE_URL/problems/upload" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{
    "id": "two-sum",
    "title": "Two Sum",
    "description": "Given an array of integers nums and an integer target, return indices of the two numbers such that they add up to target.\n\nYou may assume that each input would have exactly one solution, and you may not use the same element twice.\n\nExample:\nInput: nums = [2,7,11,15], target = 9\nOutput: [0,1]\nExplanation: Because nums[0] + nums[1] == 9, we return [0, 1].",
    "difficulty": "easy",
    "points": 10,
    "test_cases": [
      {
        "name": "Test Case 1",
        "type": "visible",
        "input": "[2,7,11,15]\n9",
        "expected_output": "[0,1]"
      },
      {
        "name": "Test Case 2",
        "type": "visible",
        "input": "[3,2,4]\n6",
        "expected_output": "[1,2]"
      },
      {
        "name": "Test Case 3",
        "type": "hidden",
        "input": "[3,3]\n6",
        "expected_output": "[0,1]"
      }
    ]
  }'
echo ""
echo ""

# Problem 2: Reverse String
echo "Adding Problem 2: Reverse String..."
curl -X POST "$BASE_URL/problems/upload" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{
    "id": "reverse-string",
    "title": "Reverse String",
    "description": "Write a function that reverses a string. The input string is given as an array of characters.\n\nYou must do this by modifying the input array in-place with O(1) extra memory.\n\nExample:\nInput: [\"h\",\"e\",\"l\",\"l\",\"o\"]\nOutput: [\"o\",\"l\",\"l\",\"e\",\"h\"]",
    "difficulty": "easy",
    "points": 10,
    "test_cases": [
      {
        "name": "Test Case 1",
        "type": "visible",
        "input": "[\"h\",\"e\",\"l\",\"l\",\"o\"]",
        "expected_output": "[\"o\",\"l\",\"l\",\"e\",\"h\"]"
      },
      {
        "name": "Test Case 2",
        "type": "hidden",
        "input": "[\"H\",\"a\",\"n\",\"n\",\"a\",\"h\"]",
        "expected_output": "[\"h\",\"a\",\"n\",\"n\",\"a\",\"H\"]"
      }
    ]
  }'
echo ""
echo ""

# Problem 3: Palindrome Number
echo "Adding Problem 3: Palindrome Number..."
curl -X POST "$BASE_URL/problems/upload" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{
    "id": "palindrome-number",
    "title": "Palindrome Number",
    "description": "Given an integer x, return true if x is a palindrome, and false otherwise.\n\nExample:\nInput: x = 121\nOutput: true\nExplanation: 121 reads as 121 from left to right and from right to left.",
    "difficulty": "easy",
    "points": 10,
    "test_cases": [
      {
        "name": "Test Case 1",
        "type": "visible",
        "input": "121",
        "expected_output": "true"
      },
      {
        "name": "Test Case 2",
        "type": "visible",
        "input": "-121",
        "expected_output": "false"
      },
      {
        "name": "Test Case 3",
        "type": "hidden",
        "input": "10",
        "expected_output": "false"
      }
    ]
  }'
echo ""
echo ""

# Problem 4: FizzBuzz
echo "Adding Problem 4: FizzBuzz..."
curl -X POST "$BASE_URL/problems/upload" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{
    "id": "fizzbuzz",
    "title": "FizzBuzz",
    "description": "Given an integer n, return a string array answer (1-indexed) where:\n- answer[i] == \"FizzBuzz\" if i is divisible by 3 and 5.\n- answer[i] == \"Fizz\" if i is divisible by 3.\n- answer[i] == \"Buzz\" if i is divisible by 5.\n- answer[i] == i (as a string) if none of the above conditions are true.\n\nExample:\nInput: n = 5\nOutput: [\"1\",\"2\",\"Fizz\",\"4\",\"Buzz\"]",
    "difficulty": "easy",
    "points": 10,
    "test_cases": [
      {
        "name": "Test Case 1",
        "type": "visible",
        "input": "5",
        "expected_output": "[\"1\",\"2\",\"Fizz\",\"4\",\"Buzz\"]"
      },
      {
        "name": "Test Case 2",
        "type": "hidden",
        "input": "15",
        "expected_output": "[\"1\",\"2\",\"Fizz\",\"4\",\"Buzz\",\"Fizz\",\"7\",\"8\",\"Fizz\",\"Buzz\",\"11\",\"Fizz\",\"13\",\"14\",\"FizzBuzz\"]"
      }
    ]
  }'
echo ""
echo ""

echo "All sample problems added successfully!"

