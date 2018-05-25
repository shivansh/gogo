# Prints 1 if n is a perfect square

# Test to check whether a given number is a perfect square.

	.data
nStr:		.asciiz "Enter n: "
n:		.word	0
i:		.word	0
isStr:		.asciiz "n is not a perfect square."
notStr:		.asciiz "n is a perfect square."

	.text


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
	# Store variables back into memory
	sw	$3, i

loop:
	lw	$3, n
	lw	$5, i
	sub	$3, $3, $5	# n -> $3
	addi	$5, $5, 2	# i -> $5
	# Store variables back into memory
	sw	$3, n
	sw	$5, i
	bgt	$3, 0, loop

	li	$2, 4
	la	$4, isStr
	syscall
	j	exit

	lw	$3, n
	# Store variables back into memory
	sw	$3, n
	bne	$3, 0, exit

	li	$2, 4
	la	$4, notStr
	syscall

exit:
	li	$2, 10
	syscall
	.end main
