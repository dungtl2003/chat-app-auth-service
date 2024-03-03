package org.service.auth.chatappauthenticationservice.service;

import lombok.AllArgsConstructor;
import org.service.auth.chatappauthenticationservice.repository.UserRepository;

@AllArgsConstructor
public class UserServiceImpl implements UserService {

	private final UserRepository userRepository;

}
