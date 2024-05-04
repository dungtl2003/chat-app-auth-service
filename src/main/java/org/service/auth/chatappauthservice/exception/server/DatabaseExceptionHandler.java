package org.service.auth.chatappauthservice.exception.server;

import org.postgresql.util.PSQLException;
import org.service.auth.chatappauthservice.exception.ErrorResponse;
import org.springframework.http.HttpStatus;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.ControllerAdvice;
import org.springframework.web.bind.annotation.ExceptionHandler;
import org.springframework.web.bind.annotation.ResponseStatus;

@ControllerAdvice
public class DatabaseExceptionHandler {

	@ExceptionHandler
	public ResponseEntity<ErrorResponse> handleException(PSQLException exception) {
		ErrorResponse error = new ErrorResponse(HttpStatus.SERVICE_UNAVAILABLE.value(),
				"There is something wrong with the server", System.currentTimeMillis());
		return new ResponseEntity<>(error, HttpStatus.SERVICE_UNAVAILABLE);
	}

}
