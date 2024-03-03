package org.service.auth.chatappauthenticationservice.exception;

import org.springframework.http.HttpStatus;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.ControllerAdvice;

@ControllerAdvice
public class UserExceptionHandler {

	public ResponseEntity<UserErrorResponse> handleException(UserNotFoundException exception) {
		UserErrorResponse error = new UserErrorResponse(HttpStatus.NOT_FOUND.value(), exception.getMessage(),
				System.currentTimeMillis());

		return new ResponseEntity<>(error, HttpStatus.NOT_FOUND);
	}

}
