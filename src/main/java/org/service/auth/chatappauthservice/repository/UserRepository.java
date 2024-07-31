package org.service.auth.chatappauthservice.repository;

import org.service.auth.chatappauthservice.entity.User;
import org.springframework.data.jpa.repository.JpaRepository;

import java.math.BigInteger;
import java.util.Optional;

public interface UserRepository extends JpaRepository<User, BigInteger> {

	Optional<User> findByEmail(String email);

}
