# Test to find sum of all even numbers less than n

	.data
nStr:		.asciiz "Enter n: "
n:		.word	0
i:		.word	0
k:		.word	0
l:		.word	0
str:		.asciiz "Sum of all even numbers less than n: "

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
	li	$3, 0		# i -> $3
	sw	$3, i		# spilled i, freed $3
	li	$3, 0		# k -> $3
	# Store dirty variables back into memory
	sw	$3, k

loop:
	lw	$3, i		# i -> $3
	lw	$5, n		# n -> $5
	bge	$3, $5, exit

	lw	$3, i		# i -> $3
	rem	$5, $3, 2
	# Store dirty variables back into memory
	sw	$5, l
	beq	$5, 1, skip

	lw	$3, k		# k -> $3
	lw	$5, i		# i -> $5
	add	$3, $3, $5
	addi	$5, $5, 1
	# Store dirty variables back into memory
	sw	$3, k
	sw	$5, i
	j	loop

skip:
	lw	$3, i		# i -> $3
	addi	$3, $3, 1
	# Store dirty variables back into memory
	sw	$3, i
	j	loop

exit:
	li	$2, 4
	la	$4, str
	syscall
	li	$2, 1
	lw	$3, k		# k -> $3
	move	$4, $3
	syscall
	li	$2, 10
	syscall
	.end main
