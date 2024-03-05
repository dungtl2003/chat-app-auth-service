package org.service.auth.chatappauthservice.DTO;

public record UserDTO(Long userId, String username, String email, String password, String role) {
}
