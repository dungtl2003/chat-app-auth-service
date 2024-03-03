package org.service.auth.chatappauthservice.exception;

public record UserErrorResponse(int status, String message, long timestamp) {
}
