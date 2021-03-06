# Prints 1 if n is a perfect square

# Test to check whether a given number is a perfect square.

	.data
nStr:		.asciiz "Enter n: "
n:		.word	0
i:		.word	0
isStr:		.asciiz "n is not a perfect square."
notStr:		.asciiz "n is a perfect square."

	.text

runtime:
	addi	$sp, $sp, -4
	sw	$ra, 0($sp)

	lw	$ra, 0($sp)
	addi	$sp, $sp, 4
	jr	$ra
	.end runtime


	.globl main
	.ent main
main:
	li	$2, 4
	la	$4, nStr
	syscall
	li	$2, 5
	syscall
	move	$3, $2
	sw	$3, n		# spilled n, freed $3
	li	$3, 1		# i -> $3
	# Store dirty variables back into memory
	sw	$3, i

loop:
	lw	$3, n		# n -> $3
	lw	$5, i		# i -> $5
	sub	$3, $3, $5
	addi	$5, $5, 2
	# Store dirty variables back into memory
	sw	$3, n
	sw	$5, i
	bgt	$3, 0, loop

	li	$2, 4
	la	$4, isStr
	syscall
	j	exit

	lw	$3, n		# n -> $3
	bne	$3, 0, exit

	li	$2, 4
	la	$4, notStr
	syscall

exit:
	li	$2, 10
	syscall
	.end main
