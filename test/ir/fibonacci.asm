# Test to find n'th fibonacci number

	.data
nStr:		.asciiz "Enter n: "
n:		.word	0
i:		.word	0
retVal:		.word	0
str:		.asciiz "n'th fibonacci number: "
first:		.word	0
second:		.word	0
temp:		.word	0

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
	li	$3, 0		# i -> $3
	sw	$3, i
	jal	fib
	lw	$3, i
	move	$3, $2
	li	$2, 4
	la	$4, str
	syscall
	li	$2, 1
	move	$4, $3
	syscall
	# Store dirty variables back into memory
	sw	$3, retVal
	li	$2, 10
	syscall
	.end main
fib:
	addi	$sp, $sp, -4
	sw	$ra, 0($sp)
	li	$3, 1		# first -> $3
	sw	$3, first	# spilled first, freed $3
	li	$3, 1		# second -> $3
	sw	$3, second	# spilled second, freed $3
	lw	$3, n		# n -> $3
	sub	$3, $3, 2
	# Store dirty variables back into memory
	sw	$3, n

loop:
	lw	$3, n		# n -> $3
	ble	$3, 0, exit

	lw	$3, second	# second -> $3
	move	$5, $3		# temp -> $5
	lw	$6, first	# first -> $6
	add	$3, $3, $6
	move	$6, $5		# first -> $6
	sw	$6, first	# spilled first, freed $6
	lw	$6, n		# n -> $6
	sub	$6, $6, 1
	# Store dirty variables back into memory
	sw	$3, second
	sw	$5, temp
	sw	$6, n
	j	loop

exit:
	lw	$2, second

	lw	$ra, 0($sp)
	addi	$sp, $sp, 4
	jr	$ra
	.end fib
