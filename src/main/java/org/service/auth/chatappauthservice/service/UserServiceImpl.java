package org.service.auth.chatappauthservice.service;

import lombok.AllArgsConstructor;
import lombok.NonNull;
import org.service.auth.chatappauthservice.entity.User;
import org.service.auth.chatappauthservice.entity.enums.TokenType;
import org.service.auth.chatappauthservice.exception.client.InvalidUserException;
import org.service.auth.chatappauthservice.exception.client.UserNotFoundException;
import org.service.auth.chatappauthservice.repository.UserRepository;
import org.springframework.security.crypto.password.PasswordEncoder;
import org.springframework.stereotype.Service;

import java.util.ArrayList;
import java.util.List;
import java.util.stream.Collectors;

@AllArgsConstructor
@Service
public class UserServiceImpl implements UserService {

	private final UserRepository userRepository;

	private final PasswordEncoder passwordEncoder;

	private final AuthTokenService authTokenService;

	@Override
	public User getValidUser(@NonNull String email, @NonNull String password)
			throws UserNotFoundException, InvalidUserException {
		User user = getUserByEmail(email);

		if (!isValid(email, password, user)) {
			throw new InvalidUserException("Invalid email and password");
		}

		return user;
	}

	@Override
	public void add(@NonNull User user) {
		user.setPassword(passwordEncoder.encode(user.getPassword()));
		userRepository.save(user);
	}

	@Override
	public void delete(@NonNull Long userId) {
		userRepository.deleteById(userId);
	}

	@Override
	public void updateUserRefreshTokens(User user, String refreshToken) {
		// Remove expired refresh tokens
		String[] refreshTokensFromDb = user.getRefreshTokens();

		List<String> refreshTokens = refreshTokensFromDb == null ? new ArrayList<>()
				: new ArrayList<>(List.of(refreshTokensFromDb)).stream()
					.filter(token -> !authTokenService.isTokenExpired(token, TokenType.REFRESH_TOKEN))
					.collect(Collectors.toList());

		refreshTokens.add(refreshToken);

		user.setRefreshTokens(refreshTokens.toArray(new String[0]));
		userRepository.save(user);
	}

	private User getUserByEmail(String email) throws UserNotFoundException {
        return userRepository
                .findByEmail(email)
                .orElseThrow(() -> new UserNotFoundException(STR."User with email \{email} does not exist"));
    }

	private boolean isValid(String email, String password, User user) {
		return email.equals(user.getEmail()) && passwordEncoder.matches(password, user.getPassword());
	}

}
